package cron

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
	"github.com/zhiyunliu/glue/transport"
)

type Server struct {
	name      string
	processor *processor
	ctx       context.Context
	opts      options
	started   bool
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
		o(&e.opts)
	}
}

func (e *Server) Name() string {
	if e.name == "" {
		e.name = e.Type()
	}
	return e.name
}

// ServiceName 服务名称
func (s *Server) ServiceName() string {
	return s.opts.serviceName
}

func (e *Server) Endpoint() *url.URL {
	return transport.NewEndpoint("cron", fmt.Sprintf("%s:%d", global.LocalIp, 1987))
}

func (e *Server) Type() string {
	return Type
}

func (e *Server) Config(cfg config.Config) {
	if cfg == nil {
		return
	}
	e.Options(WithConfig(cfg))
	cfg.Get(fmt.Sprintf("servers.%s", e.Name())).Scan(e.opts.setting)
}

// Start 开始
func (e *Server) Start(ctx context.Context) error {

	if e.opts.setting.Config.Status == server.StatusStop {
		return nil
	}

	e.ctx = transport.WithServerContext(ctx, e)
	err := e.newProcessor()
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	log.Infof("CRON Server [%s] listening on %s", e.name, global.LocalIp)

	done := make(chan struct{})
	go func() {
		e.started = true
		errChan <- e.processor.Start()
		close(done)
	}()

	select {
	case <-time.After(time.Second):
		errChan <- nil
	case <-done:
	}
	err = <-errChan
	if err != nil {
		log.Errorf("CRON Server [%s] start error: %s", e.name, err.Error())
		return err
	}
	if len(e.opts.startedHooks) > 0 {
		for _, fn := range e.opts.startedHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("CRON Server [%s] StartedHooks:%+v", e.name, err)
				return err
			}
		}
	}
	log.Infof("CRON Server [%s] start completed", e.name)
	return nil
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {

	err := e.processor.Close()
	if err != nil {
		log.Errorf("CRON Server [%s] stop error: %s", e.name, err.Error())
		return err
	}

	if len(e.opts.endHooks) > 0 {
		for _, fn := range e.opts.endHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("CRON Server [%s] EndHook:", e.name, err)
				return err
			}
		}
	}
	log.Infof("CRON Server [%s] stop completed", e.name)

	return nil
}

func (e *Server) newProcessor() error {
	var err error
	e.processor, err = newProcessor(e.opts.config)
	if err != nil {
		return err
	}

	err = e.processor.Add(e.opts.setting.Jobs...)
	if err != nil {
		return err
	}
	e.registryEngineRoute()
	return nil
}

func (e *Server) AddJob(jobs ...*Job) error {
	err := e.processor.Add(jobs...)
	if err != nil {
		return err
	}
	e.registryEngineRoute()
	return nil
}
func (e *Server) ResetRoute() {
	e.opts.router = server.NewRouterGroup("")
}
func (e *Server) RemoveJob(key string) {
	e.processor.Remove(key)
}
func (e *Server) Use(middlewares ...middleware.Middleware) {
	e.opts.router.Use(middlewares...)
}

func (e *Server) Group(group string, middlewares ...middleware.Middleware) *server.RouterGroup {
	return e.opts.router.Group(group, middlewares...)
}

func (e *Server) Handle(path string, obj interface{}) {
	e.opts.router.Handle(path, obj, server.MethodGet)
}
