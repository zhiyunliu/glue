//go:build linux || darwin

package ipc

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

func listen(endpoint string) (net.Listener, error) {
	if endpoint == "" {
		return nil, errors.New("ipc: empty endpoint")
	}

	lockFD, err := lockEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	listener, listenErr := listenUnix(endpoint)
	unlockErr := unlockEndpoint(lockFD)
	if unlockErr != nil {
		var closeErr error
		if listener != nil {
			closeErr = listener.Close()
			if closeErr != nil {
				closeErr = fmt.Errorf("ipc: close listener after startup-lock release failure: %w", closeErr)
			}
		}
		return nil, errors.Join(listenErr, fmt.Errorf("ipc: release endpoint startup lock: %w", unlockErr), closeErr)
	}
	return listener, listenErr
}

func listenUnix(endpoint string) (net.Listener, error) {
	if info, err := os.Lstat(endpoint); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, fmt.Errorf("ipc: endpoint exists and is not a socket: %s", endpoint)
		}
		conn, dialErr := net.DialTimeout("unix", endpoint, 100*time.Millisecond)
		if dialErr == nil {
			if err := conn.Close(); err != nil {
				return nil, fmt.Errorf("ipc: close socket probe %q: %w", endpoint, err)
			}
			return nil, ErrEndpointInUse
		}
		if !errors.Is(dialErr, unix.ECONNREFUSED) {
			return nil, fmt.Errorf("ipc: probe existing socket %q: %w", endpoint, dialErr)
		}
		if err := os.Remove(endpoint); err != nil {
			return nil, fmt.Errorf("ipc: remove stale socket %q: %w", endpoint, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("ipc: inspect endpoint %q: %w", endpoint, err)
	}

	listener, err := net.Listen("unix", endpoint)
	if err != nil {
		return nil, fmt.Errorf("ipc: listen on unix socket %q: %w", endpoint, err)
	}
	return listener, nil
}

func lockEndpoint(endpoint string) (int, error) {
	absoluteEndpoint, err := filepath.Abs(endpoint)
	if err != nil {
		return -1, fmt.Errorf("ipc: resolve endpoint %q: %w", endpoint, err)
	}
	endpointDir, err := filepath.EvalSymlinks(filepath.Dir(absoluteEndpoint))
	if err != nil {
		return -1, fmt.Errorf("ipc: resolve endpoint directory for %q: %w", endpoint, err)
	}
	canonicalEndpoint := filepath.Join(endpointDir, filepath.Base(absoluteEndpoint))
	hash := sha256.Sum256([]byte(canonicalEndpoint))
	lockPath := filepath.Join(endpointDir, ".ipc-listener-"+hex.EncodeToString(hash[:12])+".lock")
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
	if err != nil {
		return -1, fmt.Errorf("ipc: open endpoint startup lock %q: %w", lockPath, err)
	}

	closeOnError := func(err error) (int, error) {
		return -1, errors.Join(err, unix.Close(fd))
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return closeOnError(fmt.Errorf("ipc: inspect endpoint startup lock %q: %w", lockPath, err))
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Uid != uint32(os.Geteuid()) {
		return closeOnError(fmt.Errorf("ipc: endpoint startup lock %q is not a regular file owned by this user", lockPath))
	}
	if err := unix.Fchmod(fd, 0o600); err != nil {
		return closeOnError(fmt.Errorf("ipc: secure endpoint startup lock %q: %w", lockPath, err))
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		return closeOnError(fmt.Errorf("ipc: acquire endpoint startup lock %q: %w", lockPath, err))
	}
	return fd, nil
}

func unlockEndpoint(fd int) error {
	unlockErr := unix.Flock(fd, unix.LOCK_UN)
	closeErr := unix.Close(fd)
	if unlockErr != nil {
		unlockErr = fmt.Errorf("ipc: unlock endpoint startup lock: %w", unlockErr)
	}
	if closeErr != nil {
		closeErr = fmt.Errorf("ipc: close endpoint startup lock: %w", closeErr)
	}
	return errors.Join(unlockErr, closeErr)
}

func dialTransport(endpoint string) (net.Conn, error) {
	conn, err := net.Dial("unix", endpoint)
	if err != nil {
		return nil, fmt.Errorf("ipc: dial unix socket %q: %w", endpoint, err)
	}
	return conn, nil
}
