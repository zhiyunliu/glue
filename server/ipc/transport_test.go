package ipc

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func testEndpoint(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("glue-ipc-test-%d", time.Now().UnixNano())
	}
	return filepath.Join(t.TempDir(), "ipc.sock")
}

func TestTransportAcceptsAndExchangesData(t *testing.T) {
	endpoint := testEndpoint(t)
	listener, err := listen(endpoint)
	if err != nil {
		t.Fatalf("listen() error = %v", err)
	}
	defer listener.Close()

	type acceptResult struct {
		conn net.Conn
		err  error
	}
	accepted := make(chan acceptResult, 1)
	go func() {
		conn, err := listener.Accept()
		accepted <- acceptResult{conn: conn, err: err}
	}()

	client, err := dialTransport(endpoint)
	if err != nil {
		t.Fatalf("dialTransport() error = %v", err)
	}
	defer client.Close()

	var server net.Conn
	select {
	case result := <-accepted:
		if result.err != nil {
			t.Fatalf("Accept() error = %v", result.err)
		}
		server = result.conn
	case <-time.After(3 * time.Second):
		t.Fatal("Accept() did not return")
	}
	defer server.Close()

	want := []byte("ping")
	if _, err := client.Write(want); err != nil {
		t.Fatalf("client Write() error = %v", err)
	}
	got := make([]byte, len(want))
	if _, err := io.ReadFull(server, got); err != nil {
		t.Fatalf("server Read() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("server received %q, want %q", got, want)
	}

	if _, err := server.Write(want); err != nil {
		t.Fatalf("server Write() error = %v", err)
	}
	if _, err := io.ReadFull(client, got); err != nil {
		t.Fatalf("client Read() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("client received %q, want %q", got, want)
	}
}

func TestListenerCloseUnblocksAccept(t *testing.T) {
	listener, err := listen(testEndpoint(t))
	if err != nil {
		t.Fatalf("listen() error = %v", err)
	}

	accepted := make(chan error, 1)
	go func() {
		_, err := listener.Accept()
		accepted <- err
	}()

	if err := listener.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case err := <-accepted:
		if err == nil {
			t.Fatal("Accept() error = nil, want closed listener error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Accept() remained blocked after listener Close()")
	}
}
