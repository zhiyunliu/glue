package cron

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	_ "github.com/zhiyunliu/glue/contrib/xcron/alloter"
	_ "github.com/zhiyunliu/glue/contrib/xcron/robfigcron"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/glue/xcron"
)

type Server struct {
	name    string
	server  xcron.Server
	ctx     context.Context
	opts    options
	started bool
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
	cfg.Get(e.serverPath()).Scan(e.opts.srvCfg)
}

// Start 开始
func (e *Server) Start(ctx context.Context) (err error) {
	if e.opts.srvCfg.Config.Status == engine.StatusStop {
		return nil
	}
	e.ctx = transport.WithServerContext(ctx, e)
	e.server, err = xcron.NewServer(e.opts.srvCfg.Config.Proto,
		e.opts.router,
		e.opts.config.Get(e.serverPath()),
		engine.WithConfig(e.opts.config),
		engine.WithLogOptions(e.opts.logOpts),
		engine.WithSrvType(e.Type()),
		engine.WithSrvName(e.Name()),
		engine.WithErrorEncoder(e.opts.encErr),
		engine.WithRequestDecoder(e.opts.decReq),
		engine.WithResponseEncoder(e.opts.encResp),
	)
	if err != nil {
		return
	}

	errChan := make(chan error, 1)
	log.Infof("CRON Server [%s] listening on %s", e.name, global.LocalIp)

	done := make(chan struct{})
	go func() {
		e.started = true
		errChan <- e.server.Serve(e.ctx)
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
func (e *Server) Stop(ctx context.Context) (err error) {
	if e.server == nil {
		return
	}
	err = e.server.Stop(ctx)
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

func (e *Server) serverPath() string {
	return fmt.Sprintf("servers.%s", e.Name())
}

func (e *Server) AddJob(jobs ...*xcron.Job) (keys []string, err error) {
	if e.server == nil {
		return
	}
	keys, err = e.server.AddJob(jobs...)
	if err != nil {
		return
	}
	return
}

func (e *Server) RemoveJob(key ...string) {
	if e.server == nil {
		return
	}
	e.server.RemoveJob(key...)
}

func (e *Server) Use(middlewares ...middleware.Middleware) {
	e.opts.router.Use(middlewares...)
}

func (e *Server) Group(group string, middlewares ...middleware.Middleware) *engine.RouterGroup {
	return e.opts.router.Group(group, middlewares...)
}

func (e *Server) Handle(path string, obj interface{}) {
	e.opts.router.Handle(path, obj, engine.MethodPost)
}
