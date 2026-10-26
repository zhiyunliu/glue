//go:build linux || darwin

package ipc

import (
	"path/filepath"
	"testing"
	"time"
)

func TestEndpointStartupLockSerializesListenerStarts(t *testing.T) {
	endpoint := filepath.Join(t.TempDir(), "ipc.sock")
	firstFD, err := lockEndpoint(endpoint)
	if err != nil {
		t.Fatalf("first lockEndpoint() error = %v", err)
	}
	firstLocked := true
	t.Cleanup(func() {
		if firstLocked {
			if err := unlockEndpoint(firstFD); err != nil {
				t.Errorf("cleanup unlockEndpoint() error = %v", err)
			}
		}
	})

	started := make(chan struct{})
	acquired := make(chan int, 1)
	lockErrors := make(chan error, 1)
	go func() {
		close(started)
		fd, err := lockEndpoint(endpoint)
		if err != nil {
			lockErrors <- err
			return
		}
		acquired <- fd
	}()
	<-started

	select {
	case fd := <-acquired:
		if err := unlockEndpoint(fd); err != nil {
			t.Errorf("second unlockEndpoint() error = %v", err)
		}
		t.Fatal("second lockEndpoint() acquired the endpoint lock before release")
	case err := <-lockErrors:
		t.Fatalf("second lockEndpoint() error = %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	firstLocked = false
	if err := unlockEndpoint(firstFD); err != nil {
		t.Fatalf("first unlockEndpoint() error = %v", err)
	}
	select {
	case fd := <-acquired:
		if err := unlockEndpoint(fd); err != nil {
			t.Fatalf("second unlockEndpoint() error = %v", err)
		}
	case err := <-lockErrors:
		t.Fatalf("second lockEndpoint() error = %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("second lockEndpoint() did not acquire after release")
	}
}
