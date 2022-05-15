package mqc

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/server"
	"github.com/zhiyunliu/gel/transport"
	"github.com/zhiyunliu/golibs/xnet"
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
	e.ctx = ctx
	e.newProcessor()

	errChan := make(chan error, 1)
	log.Infof("MQC Server [%s] listening on %s", e.name, e.opts.setting.Config.String())
	go func() {
		e.started = true
		err := e.processor.Start()
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	err := <-errChan
	if err != nil {
		log.Errorf("MQC Server [%s] start error: %s", e.name, err.Error())
		return err
	}

	if len(e.opts.startedHooks) > 0 {
		for _, fn := range e.opts.startedHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("MQC Server [%s] StartedHooks:%+v", e.name, err)
				return err
			}
		}
	}
	log.Infof("MQC Server [%s] start completed", e.name)
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
		log.Errorf("MQC Server [%s] stop error: %s", e.name, err.Error())
		return err
	}

	if len(e.opts.endHooks) > 0 {
		for _, fn := range e.opts.endHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("MQC Server [%s] EndHook:", e.name, err)
				return err
			}
		}
	}
	log.Infof("MQC Server [%s] stop completed", e.name)
	return nil

}

func (e *Server) newProcessor() {
	var err error
	config := e.opts.setting.Config
	//queue://default
	protoType, configName, err := xnet.Parse(config.Addr)
	if err != nil {
		panic(err)
	}
	cfg := e.opts.config.Get(protoType)
	queueVal := cfg.Value(configName)
	ptotoCfg := queueVal.String()

	//redis://localhost
	protoType, configName, err = xnet.Parse(ptotoCfg)
	if err != nil {
		panic(err)
	}

	cfg = e.opts.config.Get(fmt.Sprintf("%s.%s", protoType, configName))
	e.processor, err = newProcessor(protoType, cfg)
	if err != nil {
		panic(err)
	}

	err = e.processor.Add(e.opts.setting.Tasks...)
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

func (e *Server) Handle(queue string, obj interface{}) {
	e.opts.router.Handle(getService(queue), obj, server.MethodGet)
}
