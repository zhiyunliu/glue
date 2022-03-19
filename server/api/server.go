package api

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"github.com/zhiyunliu/golibs/host"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/log"
	"github.com/zhiyunliu/velocity/middleware"
	"github.com/zhiyunliu/velocity/server"
	"github.com/zhiyunliu/velocity/transport"
)

type Server struct {
	name     string
	ctx      context.Context
	srv      *http.Server
	endpoint *url.URL
	opts     *options
	started  bool
}

var _ transport.Server = (*Server)(nil)

// New 实例化
func New(name string, opts ...Option) *Server {
	s := &Server{
		name: name,
		opts: setDefaultOption(),
	}
	s.Options(opts...)
	return s
}

// Options 设置参数
func (e *Server) Options(opts ...Option) {
	for _, o := range opts {
		o(e.opts)
	}
}

func (e *Server) Type() string {
	return "api"
}

func (e *Server) Name() string {
	return e.name
}

func (e *Server) Config(cfg config.Config) {
	e.Options(WithConfig(cfg))
}

// Start 开始
func (e *Server) Start(ctx context.Context) error {
	e.ctx = ctx
	e.started = true
	e.registryEngineRoute()

	lsr, err := net.Listen("tcp", e.opts.setting.Config.Addr)
	if err != nil {
		return err
	}
	err = e.setEndpoint()
	if err != nil {
		return err
	}

	e.srv = &http.Server{Handler: e.opts.handler}
	if len(e.opts.endHooks) > 0 {
		endHook := func() {
			for _, fn := range e.opts.endHooks {
				err := fn(ctx)
				if err != nil {
					log.Error("EndHook:", err)
					return
				}
			}
		}
		e.srv.RegisterOnShutdown(endHook)
	}
	e.srv.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	log.Infof("API Server [%s] listening on %s", e.name, lsr.Addr().String())
	go func() {
		if err = e.srv.Serve(lsr); err != nil {
			log.Errorf("[%s] Server start error: %s", e.name, err.Error())
		}
		<-ctx.Done()
		err = e.Stop(ctx)
		if err != nil {
			log.Errorf("[%s] Server shutdown error: %s", e.name, err.Error())
		}
	}()
	if len(e.opts.startedHooks) > 0 {
		for _, fn := range e.opts.startedHooks {
			err := fn(ctx)
			if err != nil {
				log.Error("api.StartedHooks:", err)
				return err
			}
		}
	}
	return nil
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {
	return e.srv.Shutdown(ctx)
}

//   http://127.0.0.1:8000
func (s *Server) Endpoint() *url.URL {
	return s.endpoint
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

func (e *Server) setEndpoint() error {

	addr, err := host.Extract(e.opts.setting.Config.Addr)
	if err != nil {
		return err
	}
	e.endpoint = transport.NewEndpoint("http", addr)
	return nil
}

func (e *Server) Use(middlewares ...middleware.Middleware) {
	e.opts.router.Use(middlewares...)
}

func (e *Server) Group(group string, middlewares ...middleware.Middleware) *server.RouterGroup {
	return e.opts.router.Group(group, middlewares...)
}

func (e *Server) Handle(path string, obj interface{}, methods ...server.Method) {
	e.opts.router.Handle(path, obj, methods...)
}
