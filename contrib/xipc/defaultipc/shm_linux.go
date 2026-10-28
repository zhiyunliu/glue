//go:build linux

package defaultipc

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func openSharedObject(name string, create bool) (int, bool, error) {
	path := "/dev/shm/" + strings.TrimPrefix(name, "/")
	flags := unix.O_RDWR | unix.O_CLOEXEC
	if create {
		fd, err := unix.Open(path, flags|unix.O_CREAT|unix.O_EXCL, 0600)
		if err == nil {
			return fd, true, nil
		}
		if !errors.Is(err, unix.EEXIST) {
			return -1, false, fmt.Errorf("ipc: create shared memory %q: %w", name, err)
		}
	}

	fd, err := unix.Open(path, flags, 0)
	if err != nil {
		return -1, false, fmt.Errorf("ipc: open shared memory %q: %w", name, err)
	}
	return fd, false, nil
}

func unlinkSharedObject(name string) error {
	path := "/dev/shm/" + strings.TrimPrefix(name, "/")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ipc: remove shared memory %q: %w", name, err)
	}
	return nil
}
