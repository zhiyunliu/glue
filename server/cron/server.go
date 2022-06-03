package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/server"
	"github.com/zhiyunliu/gel/transport"
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
	e.newProcessor()

	errChan := make(chan error, 1)
	log.Infof("CRON Server [%s] listening on %s", e.name, global.LocalIp)
	go func() {
		e.started = true

		done := make(chan struct{})
		go func() {
			errChan <- e.processor.Start()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-time.After(time.Second):
			errChan <- nil
		}
	}()
	err := <-errChan
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

func (e *Server) newProcessor() {
	var err error
	e.processor, err = newProcessor()
	if err != nil {
		panic(err)
	}

	err = e.processor.Add(e.opts.setting.Jobs...)
	if err != nil {
		panic(err)
	}
	e.registryEngineRoute()
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
