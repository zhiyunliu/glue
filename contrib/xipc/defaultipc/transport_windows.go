//go:build windows

package defaultipc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const pipePrefix = `\\.\pipe\`

type pipeListener struct {
	name string

	mu      sync.Mutex
	closed  bool
	pending windows.Handle
	initial windows.Handle
}

func listen(endpoint string) (net.Listener, error) {
	if endpoint == "" {
		return nil, errors.New("ipc: empty endpoint")
	}
	if len(endpoint) >= len(pipePrefix) && endpoint[:len(pipePrefix)] == pipePrefix {
		return nil, errors.New("ipc: endpoint must omit the named pipe prefix")
	}
	if _, err := windows.UTF16PtrFromString(pipePrefix + endpoint); err != nil {
		return nil, fmt.Errorf("ipc: invalid named pipe endpoint: %w", err)
	}
	name, _ := windows.UTF16PtrFromString(pipePrefix + endpoint)
	handle, err := windows.CreateNamedPipe(name, windows.PIPE_ACCESS_DUPLEX|windows.FILE_FLAG_OVERLAPPED|windows.FILE_FLAG_FIRST_PIPE_INSTANCE,
		windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT, windows.PIPE_UNLIMITED_INSTANCES, 64*1024, 64*1024, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("ipc: create named pipe listener: %w", errors.Join(ErrEndpointInUse, err))
	}
	return &pipeListener{name: pipePrefix + endpoint, initial: handle}, nil
}

func (l *pipeListener) Accept() (net.Conn, error) {
	name, err := windows.UTF16PtrFromString(l.name)
	if err != nil {
		return nil, fmt.Errorf("ipc: invalid named pipe name: %w", err)
	}
	l.mu.Lock()
	handle := l.initial
	l.initial = 0
	closed := l.closed
	l.mu.Unlock()
	if closed {
		return nil, net.ErrClosed
	}
	if handle == 0 {
		handle, err = windows.CreateNamedPipe(name, windows.PIPE_ACCESS_DUPLEX|windows.FILE_FLAG_OVERLAPPED,
			windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT, windows.PIPE_UNLIMITED_INSTANCES, 64*1024, 64*1024, 0, nil)
		if err != nil {
			return nil, fmt.Errorf("ipc: create named pipe: %w", err)
		}
	}

	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		_ = windows.CloseHandle(handle)
		return nil, net.ErrClosed
	}
	event, eventErr := windows.CreateEvent(nil, 1, 0, nil)
	if eventErr != nil {
		l.mu.Unlock()
		_ = windows.CloseHandle(handle)
		return nil, eventErr
	}
	defer windows.CloseHandle(event)
	overlapped := windows.Overlapped{HEvent: event}
	l.pending = handle
	connectErr := windows.ConnectNamedPipe(handle, &overlapped)
	l.mu.Unlock()
	if connectErr == windows.ERROR_IO_PENDING {
		var transferred uint32
		connectErr = windows.GetOverlappedResult(handle, &overlapped, &transferred, true)
	}

	l.mu.Lock()
	closed = l.closed
	if l.pending == handle {
		l.pending = 0
	}
	l.mu.Unlock()

	if closed {
		return nil, net.ErrClosed
	}
	if connectErr == windows.ERROR_NO_DATA {
		// 客户端可能在接受完成前断开，交给连接读取循环处理正常 EOF。
		return &pipeConn{handle: handle, name: l.name, serverSide: true}, nil
	}
	if connectErr != nil && connectErr != windows.ERROR_PIPE_CONNECTED {
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("ipc: connect named pipe: %w", connectErr)
	}
	return &pipeConn{handle: handle, name: l.name, serverSide: true}, nil
}

func (l *pipeListener) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	handle := l.pending
	if handle == 0 {
		handle = l.initial
	}
	l.initial = 0
	l.pending = 0
	l.mu.Unlock()
	if handle == 0 {
		return nil
	}

	cancelErr := windows.CancelIoEx(handle, nil)
	if cancelErr != nil && cancelErr != windows.ERROR_NOT_FOUND {
		cancelErr = fmt.Errorf("ipc: cancel named pipe accept: %w", cancelErr)
	} else {
		cancelErr = nil
	}
	closeErr := windows.CloseHandle(handle)
	if closeErr != nil {
		closeErr = fmt.Errorf("ipc: close pending named pipe: %w", closeErr)
	}
	return errors.Join(cancelErr, closeErr)
}

func (l *pipeListener) Addr() net.Addr {
	return &pipeAddr{name: l.name}
}

type pipeConn struct {
	handle     windows.Handle
	name       string
	serverSide bool

	closeOnce sync.Once
	closeErr  error
	ioMu      sync.Mutex
	closed    bool
}

func (c *pipeConn) Read(data []byte) (int, error)  { return c.transfer(data, false) }
func (c *pipeConn) Write(data []byte) (int, error) { return c.transfer(data, true) }

// transfer 在关闭互斥区内提交异步操作，避免取消之后再次进入阻塞读取。
func (c *pipeConn) transfer(data []byte, write bool) (int, error) {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(event)
	overlapped := windows.Overlapped{HEvent: event}
	var n uint32
	c.ioMu.Lock()
	if c.closed {
		c.ioMu.Unlock()
		return 0, net.ErrClosed
	}
	if write {
		err = windows.WriteFile(c.handle, data, &n, &overlapped)
	} else {
		err = windows.ReadFile(c.handle, data, &n, &overlapped)
	}
	c.ioMu.Unlock()
	if err == windows.ERROR_IO_PENDING {
		err = windows.GetOverlappedResult(c.handle, &overlapped, &n, true)
	}
	if err == windows.ERROR_BROKEN_PIPE || err == windows.ERROR_PIPE_NOT_CONNECTED || err == windows.ERROR_NO_DATA {
		err = io.EOF
	}
	return int(n), err
}

func (c *pipeConn) Close() error {
	c.closeOnce.Do(func() {
		c.ioMu.Lock()
		defer c.ioMu.Unlock()
		c.closed = true
		cancelErr := windows.CancelIoEx(c.handle, nil)
		if cancelErr == windows.ERROR_NOT_FOUND {
			cancelErr = nil
		}
		c.closeErr = errors.Join(cancelErr, windows.CloseHandle(c.handle))
	})
	return c.closeErr
}

func (c *pipeConn) LocalAddr() net.Addr {
	return &pipeAddr{name: c.name}
}

func (c *pipeConn) RemoteAddr() net.Addr {
	return &pipeAddr{name: c.name}
}

func (c *pipeConn) SetDeadline(time.Time) error {
	return ErrDeadlineUnsupported
}

func (c *pipeConn) SetReadDeadline(time.Time) error {
	return ErrDeadlineUnsupported
}

func (c *pipeConn) SetWriteDeadline(time.Time) error {
	return ErrDeadlineUnsupported
}

type pipeAddr struct {
	name string
}

func (a *pipeAddr) Network() string { return "pipe" }

func (a *pipeAddr) String() string { return a.name }

func dialTransport(endpoint string) (net.Conn, error) {
	if endpoint == "" {
		return nil, errors.New("ipc: empty endpoint")
	}
	name := pipePrefix + endpoint
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("ipc: invalid named pipe endpoint: %w", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	waitNamedPipe := windows.NewLazySystemDLL("kernel32.dll").NewProc("WaitNamedPipeW")
	for {
		handle, err := windows.CreateFile(
			namePtr,
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_FLAG_OVERLAPPED,
			0,
		)
		if err == nil {
			return &pipeConn{handle: handle, name: name}, nil
		}
		if err != windows.ERROR_FILE_NOT_FOUND && err != windows.ERROR_PIPE_BUSY {
			return nil, fmt.Errorf("ipc: dial named pipe: %w", err)
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("ipc: dial named pipe %q: %w", name, err)
		}

		if err == windows.ERROR_PIPE_BUSY {
			result, _, waitErr := waitNamedPipe.Call(
				uintptr(unsafe.Pointer(namePtr)),
				uintptr(100),
			)
			runtime.KeepAlive(namePtr)
			if result == 0 && waitErr != windows.ERROR_SEM_TIMEOUT && waitErr != windows.ERROR_FILE_NOT_FOUND {
				return nil, fmt.Errorf("ipc: wait for named pipe: %w", waitErr)
			}
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

var _ net.Conn = (*pipeConn)(nil)
var _ net.Listener = (*pipeListener)(nil)
