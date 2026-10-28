//go:build linux || darwin

package defaultipc

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

type unixShared struct {
	data      []byte
	closeOnce sync.Once
	closeErr  error
}

func openShared(name string, size int, create bool) (SharedRegion, error) {
	if size <= 0 {
		return nil, errors.New("ipc: shared memory size must be positive")
	}
	shmName, err := normalizeSharedName(name)
	if err != nil {
		return nil, err
	}

	fd, created, err := openSharedObject(shmName, create)
	if err != nil {
		return nil, err
	}
	closeFD := func() error {
		if err := unix.Close(fd); err != nil {
			return fmt.Errorf("ipc: close shared memory descriptor %q: %w", shmName, err)
		}
		return nil
	}

	if created {
		if err := unix.Ftruncate(fd, int64(size)); err != nil {
			unlinkErr := unlinkSharedObject(shmName)
			return nil, errors.Join(
				fmt.Errorf("ipc: size shared memory %q: %w", shmName, err),
				unlinkErr,
				closeFD(),
			)
		}
	}

	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return nil, errors.Join(
			fmt.Errorf("ipc: inspect shared memory %q: %w", shmName, err),
			closeFD(),
		)
	}
	if stat.Size < int64(size) {
		return nil, errors.Join(
			fmt.Errorf("ipc: shared memory %q is %d bytes, requested %d", shmName, stat.Size, size),
			closeFD(),
		)
	}

	data, err := unix.Mmap(fd, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("ipc: map shared memory %q: %w", shmName, err),
			closeFD(),
		)
	}
	if err := closeFD(); err != nil {
		return nil, errors.Join(err, unix.Munmap(data))
	}
	return &unixShared{data: data}, nil
}

func normalizeSharedName(name string) (string, error) {
	name = strings.TrimPrefix(name, "/")
	if name == "" || strings.Contains(name, "/") || strings.ContainsRune(name, '\x00') {
		return "", errors.New("ipc: invalid shared memory name")
	}
	return "/" + name, nil
}

func (s *unixShared) Bytes() []byte {
	return s.data
}

func (s *unixShared) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = unix.Munmap(s.data)
	})
	return s.closeErr
}
