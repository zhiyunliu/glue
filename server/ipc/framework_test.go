package ipc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/config/file"
	gctx "github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/transport"
	_ "github.com/zhiyunliu/xbinding/binding"
)

var _ transport.Server = (*Server)(nil)
var _ transport.Endpointer = (*Server)(nil)

func TestFrameworkLifecycle(t *testing.T) {
	s := New("local", WithAddr(testEndpoint(t)), WithServiceName("worker"))
	_ = glue.Server(s)
	if s.listener != nil {
		t.Fatal("New listened before Start")
	}
	if s.Name() != "local" || s.Type() != "ipc" || s.ServiceName() != "worker" {
		t.Fatal("invalid framework identity")
	}
	if s.Endpoint().Scheme != "ipc" {
		t.Fatal("invalid endpoint scheme")
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.RegisterFunc("Echo", func(r Request) *Response {
		var v echoArgs
		if err := r.Bind(&v); err != nil {
			return WrapError(CodeBadRequest, "bad_args", err)
		}
		return OK(v)
	})
	for cycle := 0; cycle < 3; cycle++ {
		if err := s.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := s.Start(context.Background()); !errors.Is(err, ErrServerAlreadyServing) {
			t.Fatalf("duplicate Start = %v", err)
		}
		c, err := Dial(s.Address())
		if err != nil {
			t.Fatal(err)
		}
		for _, value := range []string{"", "hello", strings.Repeat("x", 4096)} {
			r, err := c.Call("Echo", echoArgs{Value: value})
			if err != nil {
				t.Fatal(err)
			}
			var got echoArgs
			if err := r.Bind(&got); err != nil || !r.IsOK() || got.Value != value {
				t.Fatalf("response = %#v, %v", r, err)
			}
		}
		if err := s.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Call("Echo", nil); err == nil {
			t.Fatal("connection survived Stop")
		}
		_ = c.Close()
		if err := s.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFrameworkConfigAndErrors(t *testing.T) {
	s := New("")
	if s.Name() != "ipc" {
		t.Fatal(s.Name())
	}
	if err := s.Config(nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	data := fmt.Sprintf(`{"servers":{"ipc":{"config":{"addr":%q,"engine":"gin","status":"stop","stop_maximum_timeout":2}}}}`, testEndpoint(t))
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.New(config.WithSource(file.NewSource(path)))
	defer cfg.Close()
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}
	if err := s.Config(cfg); err != nil {
		t.Fatal(err)
	}
	if s.StopMaximumTimeout() != 2*time.Second {
		t.Fatal(s.StopMaximumTimeout())
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.listener != nil {
		t.Fatal("disabled server listened")
	}
	if err := New("missing").Config(cfg); err == nil {
		t.Fatal("missing config accepted")
	}
	for _, s := range []*Server{New("bad", WithAddr("")), New("bad", WithAddr(testEndpoint(t)), WithEngine("missing"))} {
		if err := s.Start(context.Background()); err == nil {
			_ = s.Close()
			t.Fatal("invalid configuration accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := New("cancelled", WithAddr(testEndpoint(t))).Start(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start = %v", err)
	}
}

func TestFrameworkRoutesAndMiddleware(t *testing.T) {
	s := New("routes", WithAddr(testEndpoint(t)))
	var seen []string
	mark := func(name string) middleware.Middleware {
		return func(next middleware.Handler) middleware.Handler {
			return func(ctx gctx.Context) interface{} { seen = append(seen, name); return next(ctx) }
		}
	}
	s.Use(mark("root"))
	group := s.Group("/v1", mark("group"))
	group.Handle("/echo", func(ctx gctx.Context) interface{} {
		seen = append(seen, "handler")
		if ctx.ServerType() != "ipc" || ctx.ServerName() != "routes" {
			return errors.New("invalid server context")
		}
		if server, ok := transport.FromServerContext(ctx.Context()); !ok || server != s {
			return errors.New("missing transport context")
		}
		var args echoArgs
		if err := ctx.Bind(&args); err != nil {
			return err
		}
		return args
	}, engine.MethodPost)
	s.Handle("/fail", func(gctx.Context) interface{} { return errors.New("business failure") })
	s.RegisterFunc("Legacy", func(Request) *Response { return OK("legacy") })
	if paths := s.RouterPathList().GetPathList(); len(paths) != 3 {
		t.Fatalf("paths = %v", paths)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := Dial(s.Address())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, value := range []string{"", "normal", strings.Repeat("z", 2048)} {
		seen = nil
		response, err := c.Call("/v1/echo", echoArgs{Value: value})
		if err != nil {
			t.Fatal(err)
		}
		var got echoArgs
		if err := response.Bind(&got); err != nil || !response.IsOK() || got.Value != value {
			t.Fatalf("response = %#v, %v", response, err)
		}
		if strings.Join(seen, ",") != "root,group,handler" {
			t.Fatalf("middleware order = %v", seen)
		}
	}
	for _, tc := range []struct {
		cmd  string
		data any
		code int
	}{{"/missing", nil, 404}, {"/fail", nil, 500}, {"/v1/echo", 42, 500}, {"Legacy", nil, 0}} {
		r, err := c.Call(tc.cmd, tc.data)
		if err != nil || r.Code != tc.code {
			t.Fatalf("%s = %#v, %v", tc.cmd, r, err)
		}
	}
}

func TestFrameworkStopDeadline(t *testing.T) {
	s := New("blocked", WithAddr(testEndpoint(t)), WithStopMaximumTimeout(1))
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
	callDone := make(chan error, 1)
	go func() { _, err := c.Call("Block", nil); callDone <- err }()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := s.Stop(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Stop = %v", err)
	}
	if err := s.Start(context.Background()); !errors.Is(err, ErrServerAlreadyServing) {
		t.Fatalf("Start while draining = %v", err)
	}
	close(release)
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-callDone:
		if err == nil {
			t.Fatal("blocked connection was not closed")
		}
	case <-time.After(time.Second):
		t.Fatal("call still blocked")
	}
}

func TestFrameworkOccupiedEndpoint(t *testing.T) {
	addr := testEndpoint(t)
	first := New("first", WithAddr(addr))
	if err := first.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second := New("second", WithAddr(addr))
	if err := second.Start(context.Background()); err == nil {
		_ = second.Close()
		t.Fatal("occupied endpoint accepted")
	}
}
