//go:build windows

package defaultipc

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

type winShared struct {
	data      []byte
	addr      uintptr
	handle    windows.Handle
	closeOnce sync.Once
	closeErr  error
}

//go:nocheckptr
func openShared(name string, size int, create bool) (SharedRegion, error) {
	if name == "" || size <= 0 {
		return nil, errors.New("ipc: shared memory name and size must be non-empty and positive")
	}
	size64 := uint64(size)

	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("ipc: invalid shared memory name: %w", err)
	}

	var handle windows.Handle
	if create {
		handle, err = windows.CreateFileMapping(
			windows.InvalidHandle,
			nil,
			windows.PAGE_READWRITE,
			uint32(size64>>32),
			uint32(size64),
			namePtr,
		)
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			err = nil
		}
	} else {
		handle, err = openFileMapping(namePtr)
	}
	if err != nil {
		closeErr := error(nil)
		if handle != 0 {
			closeErr = windows.CloseHandle(handle)
		}
		return nil, errors.Join(fmt.Errorf("ipc: open shared memory %q: %w", name, err), closeErr)
	}

	addr, err := windows.MapViewOfFile(
		handle,
		windows.FILE_MAP_READ|windows.FILE_MAP_WRITE,
		0,
		0,
		uintptr(size),
	)
	if err != nil {
		closeErr := windows.CloseHandle(handle)
		return nil, errors.Join(fmt.Errorf("ipc: map shared memory %q: %w", name, err), closeErr)
	}
	if addr == 0 {
		closeErr := windows.CloseHandle(handle)
		return nil, errors.Join(errors.New("ipc: map shared memory returned a null address"), closeErr)
	}
	data := mappedBytes(addr, size)
	return &winShared{data: data, addr: addr, handle: handle}, nil
}

// mappedBytes 将 Windows 映射 API 返回的地址包装为字节切片。
//
// nolint
//
//go:nocheckptr
func mappedBytes(addr uintptr, size int) []byte {
	//nolint
	return unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
}

func openFileMapping(name *uint16) (windows.Handle, error) {
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("OpenFileMappingW")
	handle, _, callErr := proc.Call(
		uintptr(windows.FILE_MAP_READ|windows.FILE_MAP_WRITE),
		0,
		uintptr(unsafe.Pointer(name)),
	)
	runtime.KeepAlive(name)
	if handle == 0 {
		if callErr == nil {
			callErr = windows.ERROR_INVALID_HANDLE
		}
		return 0, callErr
	}
	return windows.Handle(handle), nil
}

func (s *winShared) Bytes() []byte {
	return s.data
}

func (s *winShared) Close() error {
	s.closeOnce.Do(func() {
		unmapErr := windows.UnmapViewOfFile(s.addr)
		closeErr := windows.CloseHandle(s.handle)
		s.closeErr = errors.Join(unmapErr, closeErr)
	})
	return s.closeErr
}
