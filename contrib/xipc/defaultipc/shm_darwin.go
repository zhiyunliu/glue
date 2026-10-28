//go:build darwin

package defaultipc

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

func openSharedObject(name string, create bool) (int, bool, error) {
	namePtr, err := unix.BytePtrFromString(name)
	if err != nil {
		return -1, false, fmt.Errorf("ipc: invalid shared memory name: %w", err)
	}
	flags := uintptr(unix.O_RDWR)
	if create {
		fd, callErr := syscallShmOpen(namePtr, flags|uintptr(unix.O_CREAT|unix.O_EXCL), 0600)
		if callErr == nil {
			return fd, true, nil
		}
		if callErr != unix.EEXIST {
			return -1, false, fmt.Errorf("ipc: create shared memory %q: %w", name, callErr)
		}
	}

	fd, callErr := syscallShmOpen(namePtr, flags, 0)
	if callErr != nil {
		return -1, false, fmt.Errorf("ipc: open shared memory %q: %w", name, callErr)
	}
	return fd, false, nil
}

func syscallShmOpen(name *byte, flags uintptr, mode uintptr) (int, error) {
	fd, _, errno := unix.Syscall(unix.SYS_SHM_OPEN, uintptr(unsafe.Pointer(name)), flags, mode)
	runtime.KeepAlive(name)
	if errno != 0 {
		return -1, errno
	}
	return int(fd), nil
}

func unlinkSharedObject(name string) error {
	namePtr, err := unix.BytePtrFromString(name)
	if err != nil {
		return fmt.Errorf("ipc: invalid shared memory name: %w", err)
	}
	_, _, errno := unix.Syscall(unix.SYS_SHM_UNLINK, uintptr(unsafe.Pointer(namePtr)), 0, 0)
	runtime.KeepAlive(namePtr)
	if errno != 0 && errno != unix.ENOENT {
		return fmt.Errorf("ipc: unlink shared memory %q: %w", name, errno)
	}
	return nil
}
