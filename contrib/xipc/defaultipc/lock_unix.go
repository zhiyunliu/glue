//go:build linux || darwin

package defaultipc

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

type unixLock struct {
	file      *os.File
	closeOnce sync.Once
	closeErr  error
}

func openLock(name string) (ProcessLock, error) {
	if name == "" {
		return nil, fmt.Errorf("ipc: empty lock name")
	}
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("ipc: open lock %q: %w", name, err)
	}
	return &unixLock{file: file}, nil
}

func (l *unixLock) Lock() error {
	if err := unix.Flock(int(l.file.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("ipc: acquire lock: %w", err)
	}
	return nil
}

func (l *unixLock) Unlock() error {
	if err := unix.Flock(int(l.file.Fd()), unix.LOCK_UN); err != nil {
		return fmt.Errorf("ipc: release lock: %w", err)
	}
	return nil
}

func (l *unixLock) Close() error {
	l.closeOnce.Do(func() {
		l.closeErr = l.file.Close()
	})
	return l.closeErr
}
