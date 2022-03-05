package api

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhiyunliu/velocity/libs/types"
	log "github.com/zhiyunliu/velocity/logger"
	"github.com/zhiyunliu/velocity/server"
)

type Server struct {
	name    string
	ctx     context.Context
	srv     *http.Server
	opts    options
	started bool
}

// New 实例化
func New(name string, opts ...Option) server.Runnable {
	s := &Server{
		name: name,
		opts: setDefaultOption(),
	}
	s.Options(opts...)
	return s
}

// NewMetrics 新建默认监控服务
func NewMetrics(opts ...Option) server.Runnable {
	s := &Server{
		name: "metrics",
		opts: setDefaultOption(),
	}
	s.opts.addr = ":3000"
	h := http.NewServeMux()
	h.Handle("/metrics", promhttp.Handler())
	s.opts.handler = h
	s.Options(opts...)
	return s
}

// NewHealthz 默认健康检查服务
func NewHealthz(opts ...Option) server.Runnable {
	s := &Server{
		name: "health",
		opts: setDefaultOption(),
	}
	s.opts.addr = ":4000"
	h := http.NewServeMux()
	h.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s.opts.handler = h
	s.Options(opts...)
	return s
}

// Options 设置参数
func (e *Server) Options(opts ...Option) {
	for _, o := range opts {
		o(&e.opts)
	}
}

func (e *Server) String() string {
	return e.name
}

// Start 开始
func (e *Server) Start() error {
	ctx := context.Background()
	l, err := net.Listen("tcp", e.opts.addr)
	if err != nil {
		return err
	}
	e.ctx = ctx
	e.started = true
	e.srv = &http.Server{Handler: e.opts.handler}
	if e.opts.endHook != nil {
		e.srv.RegisterOnShutdown(e.opts.endHook)
	}
	e.srv.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
	log.Infof("%s Server listening on %s", e.name, l.Addr().String())
	go func() {
		if err = e.srv.Serve(l); err != nil {
			log.Errorf("%s Server start error: %s", e.name, err.Error())
		}
		<-ctx.Done()
		err = e.Shutdown(ctx)
		if err != nil {
			log.Errorf("%S Server shutdown error: %s", e.name, err.Error())
		}
	}()
	if e.opts.startedHook != nil {
		e.opts.startedHook()
	}
	return nil
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

// Shutdown 停止
func (e *Server) Shutdown(ctx context.Context) error {
	return e.srv.Shutdown(ctx)
}

// Shutdown 停止
func (e *Server) Stop() error {
	return nil
}

// Shutdown 停止
func (e *Server) Status() string {
	return ""
}

func (e *Server) Port() uint64 {
	return 0
}

func (e *Server) Metadata() types.XMap {
	return types.XMap{}
}
func (e *Server) Name() string {
	return e.name
}
