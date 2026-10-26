package ipc

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	gctx "github.com/zhiyunliu/glue/context"
)

func TestFrameworkHooksRollbackAndRestart(t *testing.T) {
	failure := errors.New("startup hook failed")
	var starts, ends atomic.Int32
	s := New("hooks", WithAddr(testEndpoint(t)), WithStartedHook(func(context.Context) error {
		if starts.Add(1) == 1 {
			return failure
		}
		return nil
	}), WithEndHook(func(context.Context) error { ends.Add(1); return nil }))
	if err := s.Start(context.Background()); !errors.Is(err, failure) {
		t.Fatalf("Start = %v", err)
	}
	if !s.Attempt() || ends.Load() != 1 {
		t.Fatal("failed startup was not rolled back")
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if starts.Load() != 2 || ends.Load() != 2 {
		t.Fatalf("hooks = %d, %d", starts.Load(), ends.Load())
	}
}

func TestFrameworkStopMaximumTimeout(t *testing.T) {
	s := New("timeout", WithAddr(testEndpoint(t)), WithStopMaximumTimeout(1))
	entered, release := make(chan struct{}), make(chan struct{})
	s.RegisterFunc("Block", func(Request) *Response { close(entered); <-release; return OK(nil) })
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	c, err := Dial(s.Address())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	done := make(chan struct{})
	go func() { defer close(done); _, _ = c.Call("Block", nil) }()
	<-entered
	start := time.Now()
	err = s.Stop(context.Background())
	elapsed := time.Since(start)
	close(release)
	if !errors.Is(err, context.DeadlineExceeded) || elapsed > 3*time.Second {
		t.Fatalf("Stop = %v after %v", err, elapsed)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	<-done
}

type failedListener struct {
	net.Listener
	err error
}

func (l failedListener) Accept() (net.Conn, error) { return nil, l.err }

func TestFrameworkServePropagatesAcceptError(t *testing.T) {
	s, err := NewServer(testEndpoint(t))
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("accept failed")
	s.command.listener = failedListener{Listener: s.command.listener, err: failure}
	if err := s.Serve(); !errors.Is(err, failure) {
		t.Fatalf("Serve = %v", err)
	}
	if err := s.Stop(context.Background()); !errors.Is(err, failure) {
		t.Fatalf("Stop = %v", err)
	}
}

type frameworkObject struct{}

func (frameworkObject) EchoHandle(ctx gctx.Context) interface{} {
	var args echoArgs
	if err := ctx.Bind(&args); err != nil {
		return err
	}
	return OK(args)
}

func TestFrameworkObjectRoutesAcrossRestart(t *testing.T) {
	s := New("object", WithAddr(testEndpoint(t)))
	s.Group("/v2").Handle("/object", frameworkObject{})
	for i := 0; i < 3; i++ {
		if err := s.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		c, err := Dial(s.Address())
		if err != nil {
			t.Fatal(err)
		}
		r, err := c.Call("/v2/object/echo", echoArgs{Value: "object"})
		if err != nil {
			t.Fatal(err)
		}
		var got echoArgs
		if err := r.Bind(&got); err != nil || !r.IsOK() || got.Value != "object" {
			t.Fatalf("response = %#v, %v", r, err)
		}
		_ = c.Close()
		if err := s.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFrameworkCloseIsTerminal(t *testing.T) {
	s := New("closed", WithAddr(testEndpoint(t)))
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := s.Start(context.Background()); !errors.Is(err, net.ErrClosed) {
			t.Fatalf("Start after Close = %v", err)
		}
		if err := s.Serve(); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
