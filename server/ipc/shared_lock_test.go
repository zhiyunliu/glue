package ipc

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func testSharedName() string {
	return fmt.Sprintf("glue_ipc_test_%d_%d", time.Now().UnixNano(), time.Now().UnixNano()%997)
}

func testLockName(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("glue-ipc-lock-%d", time.Now().UnixNano())
	}
	return filepath.Join(t.TempDir(), "ipc.lock")
}

func TestOpenSharedRejectsInvalidSize(t *testing.T) {
	if _, err := OpenShared(testSharedName(), 0, true); err == nil {
		t.Fatal("OpenShared() error = nil, want error for zero size")
	}
}

func TestOpenSharedPreservesExistingData(t *testing.T) {
	name := testSharedName()
	t.Cleanup(func() {
		if err := cleanupShared(name); err != nil {
			t.Errorf("cleanupShared() error = %v", err)
		}
	})

	first, err := OpenShared(name, 8, true)
	if err != nil {
		t.Fatalf("OpenShared(create) error = %v", err)
	}
	copy(first.Bytes(), []byte("preserve"))

	second, err := OpenShared(name, 8, true)
	if err != nil {
		t.Fatalf("OpenShared(existing) error = %v", err)
	}
	if got := string(second.Bytes()); got != "preserve" {
		t.Fatalf("existing shared bytes = %q, want %q", got, "preserve")
	}
	if err := second.Close(); err != nil {
		t.Errorf("second Close() error = %v", err)
	}
	if err := first.Close(); err != nil {
		t.Errorf("first Close() error = %v", err)
	}
}

func TestProcessLockCanAcquireAndRelease(t *testing.T) {
	lock, err := OpenLock(testLockName(t))
	if err != nil {
		t.Fatalf("OpenLock() error = %v", err)
	}
	if err := lock.Lock(); err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	if err := lock.Unlock(); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
