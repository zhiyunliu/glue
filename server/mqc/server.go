package mqc

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	_ "github.com/zhiyunliu/glue/contrib/xmqc/alloter"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/glue/xmqc"
)

const (
	Port = 1988
)

type Server struct {
	name    string
	server  xmqc.Server
	ctx     context.Context
	opts    *options
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
	return transport.NewEndpoint(e.Type(), fmt.Sprintf("%s:%d", global.LocalIp, Port))
}

func (e *Server) Config(cfg config.Config) {
	if cfg == nil {
		return
	}
	e.Options(WithConfig(cfg))
	cfg.Get(e.serverPath()).ScanTo(e.opts.srvCfg)
}

// Start 开始
func (e *Server) Start(ctx context.Context) (err error) {
	if e.opts.srvCfg.Config.Status == engine.StatusStop {
		return nil
	}

	e.ctx = transport.WithServerContext(ctx, e)

	e.server, err = xmqc.NewServer(e.opts.srvCfg.Config.Proto,
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
	log.Infof("MQC Server [%s] listening on %s", e.name, e.opts.srvCfg.Config.Addr)
	done := make(chan struct{})

	go func() {
		e.started = true
		serveErr := e.server.Serve(e.ctx) //存在1s内，服务没有启动的可能性
		if serveErr != nil {
			log.Errorf("MQC Server [%s] Serve error: %s", e.name, serveErr.Error())
		}
		errChan <- serveErr
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		errChan <- nil
	}
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

// 获取树形的路径列表
func (s *Server) RouterPathList() transport.RouterList {
	return engine.RouterList{
		ServerType: s.Type(),
		PathList:   s.opts.router.GetTreePathList(),
	}
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {
	if e.server == nil {
		return nil
	}
	err := e.server.Stop(ctx)
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

func (e *Server) Use(middlewares ...middleware.Middleware) {
	e.opts.router.Use(middlewares...)
}

func (e *Server) Group(group string, middlewares ...middleware.Middleware) *engine.RouterGroup {
	return e.opts.router.Group(group, middlewares...)
}

func (e *Server) Handle(queue string, obj interface{}, opts ...engine.RouterOption) {
	newopts := append(opts, engine.MethodPost)
	e.opts.router.Handle(xmqc.GetService(queue), obj, newopts...)
}

func (e *Server) serverPath() string {
	return fmt.Sprintf("servers.%s", e.Name())
}
