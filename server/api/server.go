package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/golibs/xnet"
)

type Server struct {
	ctx      context.Context
	name     string
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
	if e.name == "" {
		e.name = e.Type()
	}
	return e.name
}

func (e *Server) Config(cfg config.Config) {
	if cfg == nil {
		return
	}
	e.Options(WithConfig(cfg))
	cfg.Get(fmt.Sprintf("servers.%s", e.Name())).Scan(e.opts.setting)
}

// Start 开始
func (e *Server) Start(ctx context.Context) (err error) {
	if e.opts.setting.Config.Status == server.StatusStop {
		return nil
	}
	e.opts.setting.Config.Addr, err = xnet.GetAvaliableAddr(log.DefaultLogger, global.LocalIp, e.opts.setting.Config.Addr)
	if err != nil {
		return err
	}

	e.ctx = transport.WithServerContext(ctx, e)
	e.started = true
	e.registryEngineRoute()

	lsr, err := net.Listen("tcp", e.opts.setting.Config.Addr)
	if err != nil {
		return err
	}

	e.srv = &http.Server{
		Handler:           e.opts.handler,
		ReadTimeout:       time.Duration(e.opts.setting.Config.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(e.opts.setting.Config.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(e.opts.setting.Config.WriteTimeout) * time.Second,
		MaxHeaderBytes:    int(e.opts.setting.Config.MaxHeaderBytes),
	}
	if len(e.opts.endHooks) > 0 {
		endHook := func() {
			for _, fn := range e.opts.endHooks {
				err := fn(ctx)
				if err != nil {
					log.Errorf("API Server [%s] EndHook:%+v", e.name, err)
					return
				}
			}
		}
		e.srv.RegisterOnShutdown(endHook)
	}
	e.srv.BaseContext = func(_ net.Listener) context.Context {
		return e.ctx
	}
	log.Infof("API Server [%s] listening on %s", e.name, e.opts.setting.Config.Addr)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		errChan <- e.srv.Serve(lsr)
		close(done)
	}()

	select {
	case <-time.After(time.Second):
		errChan <- nil
	case <-done:
	}
	err = <-errChan
	if err != nil {
		log.Errorf("API Server [%s] start error: %s", e.name, err.Error())
		return err
	}
	if len(e.opts.startedHooks) > 0 {
		for _, fn := range e.opts.startedHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("API Server [%s] StartedHooks:%+v", e.name, err)
				return err
			}
		}
	}

	log.Infof("API Server [%s] start completed", e.name)
	return nil
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {
	e.started = false
	err := e.srv.Shutdown(ctx)
	if err != nil {
		log.Errorf("API Server [%s] stop error: %s", e.name, err.Error())
		return err
	}
	log.Infof("API Server [%s] stop completed", e.name)
	return err
}

// ServiceName 服务名称
func (s *Server) ServiceName() string {
	return s.opts.serviceName
}

// http://127.0.0.1:8000
func (s *Server) Endpoint() *url.URL {
	if s.endpoint == nil {
		s.endpoint = s.buildEndpoint()
	}
	return s.endpoint
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

func (e *Server) buildEndpoint() *url.URL {
	host, port, err := xnet.ExtractHostPort(e.opts.setting.Config.Addr)
	if err != nil {
		panic(fmt.Errorf("API Server Addr:%s 配置错误", e.opts.setting.Config.Addr))
	}
	if host == "" {
		host = global.LocalIp
	}
	return transport.NewEndpoint("http", fmt.Sprintf("%s:%d", host, port))
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

func (e *Server) StaticFile(path, filepath string) {
	e.opts.static[path] = Static{RouterPath: path, FilePath: filepath, IsFile: true}
}

func (e *Server) Static(path, root string) {
	e.opts.static[path] = Static{RouterPath: path, FilePath: root, IsFile: false}
}
