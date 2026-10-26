//go:build windows

package ipc

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"golang.org/x/sys/windows"
)

type winLock struct {
	handle windows.Handle

	gate       sync.Mutex
	mu         sync.Mutex
	acquiring  bool
	locked     bool
	closed     bool
	closeOnce  sync.Once
	closeError error
}

func openLock(name string) (ProcessLock, error) {
	if name == "" {
		return nil, errors.New("ipc: empty lock name")
	}
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("ipc: invalid lock name: %w", err)
	}

	handle, err := windows.OpenMutex(windows.MUTEX_ALL_ACCESS, false, namePtr)
	if err != nil {
		if err != windows.ERROR_FILE_NOT_FOUND {
			return nil, fmt.Errorf("ipc: open mutex %q: %w", name, err)
		}
		handle, err = windows.CreateMutex(nil, false, namePtr)
		if err != nil && err != windows.ERROR_ALREADY_EXISTS {
			closeErr := error(nil)
			if handle != 0 {
				closeErr = windows.CloseHandle(handle)
			}
			return nil, errors.Join(fmt.Errorf("ipc: create mutex %q: %w", name, err), closeErr)
		}
	}
	return &winLock{handle: handle}, nil
}

func (l *winLock) Lock() error {
	l.gate.Lock()

	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		l.gate.Unlock()
		return errors.New("ipc: lock is closed")
	}
	l.acquiring = true
	l.mu.Unlock()

	runtime.LockOSThread()
	result, err := windows.WaitForSingleObject(l.handle, windows.INFINITE)

	l.mu.Lock()
	l.acquiring = false
	acquired := result == windows.WAIT_OBJECT_0 || result == windows.WAIT_ABANDONED
	if acquired {
		l.locked = true
	}
	l.mu.Unlock()

	if !acquired {
		runtime.UnlockOSThread()
		l.gate.Unlock()
		if err != nil {
			return fmt.Errorf("ipc: wait for mutex: %w", err)
		}
		return fmt.Errorf("ipc: unexpected mutex wait result: %#x", result)
	}
	return nil
}

func (l *winLock) Unlock() error {
	l.mu.Lock()
	if !l.locked {
		l.mu.Unlock()
		return errors.New("ipc: mutex is not locked")
	}
	l.mu.Unlock()

	err := windows.ReleaseMutex(l.handle)
	if err != nil {
		return fmt.Errorf("ipc: release mutex: %w", err)
	}
	l.mu.Lock()
	l.locked = false
	l.mu.Unlock()
	runtime.UnlockOSThread()
	l.gate.Unlock()
	return nil
}

func (l *winLock) Close() error {
	l.mu.Lock()
	if l.acquiring || l.locked {
		l.mu.Unlock()
		return errors.New("ipc: cannot close a locked mutex")
	}
	if l.closed {
		l.mu.Unlock()
		return l.closeError
	}
	l.closed = true
	l.mu.Unlock()

	l.closeOnce.Do(func() {
		l.closeError = windows.CloseHandle(l.handle)
	})
	if l.closeError != nil {
		return fmt.Errorf("ipc: close mutex: %w", l.closeError)
	}
	return nil
}
