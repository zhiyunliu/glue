package mqc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/golibs/xnet"
)

type Server struct {
	name      string
	processor *processor
	ctx       context.Context
	opts      *options
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
		o(e.opts)
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

// ServiceName 服务名称
func (s *Server) ServiceName() string {
	return s.opts.serviceName
}

func (e *Server) Endpoint() *url.URL {
	return transport.NewEndpoint("mqc", fmt.Sprintf("%s:%d", global.LocalIp, 1987))
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

	e.ctx = transport.WithServerContext(ctx, e)
	err = e.newProcessor()
	if err != nil {
		return
	}

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

	err = <-errChan
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

func (e *Server) newProcessor() (err error) {
	config := e.opts.setting.Config
	//queue://default
	protoType, configName, err := xnet.Parse(config.Addr)
	if err != nil {
		return
	}
	//{"proto":"redis","addr":"redis://localhost"}
	cfg := e.opts.config.Get(protoType).Get(configName)

	protoType = cfg.Value("proto").String()

	e.processor, err = newProcessor(e.ctx, protoType, cfg)
	if err != nil {
		return
	}

	err = e.processor.Add(e.opts.setting.Tasks...)
	if err != nil {
		return
	}
	err = e.resoverEngineRoute(e.processor)
	return
}

func (e *Server) Use(middlewares ...middleware.Middleware) {
	e.opts.router.Use(middlewares...)
}

func (e *Server) Group(group string, middlewares ...middleware.Middleware) *engine.RouterGroup {
	return e.opts.router.Group(group, middlewares...)
}

func (e *Server) Handle(queue string, obj interface{}) {
	e.opts.router.Handle(getService(queue), obj, engine.MethodGet)
}
