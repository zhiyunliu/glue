package rpc

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	_ "github.com/zhiyunliu/glue/contrib/xrpc/grpc"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/xnet"
)

type Server struct {
	ctx      context.Context
	name     string
	server   xrpc.Server
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
	cfg.Get(e.serverPath()).Scan(e.opts.srvCfg)
}

// Start 开始
func (e *Server) Start(ctx context.Context) (err error) {
	if e.opts.srvCfg.Config.Status == engine.StatusStop {
		return nil
	}
	e.ctx = transport.WithServerContext(ctx, e)

	e.server, err = xrpc.NewServer(e.opts.srvCfg.Config.Proto,
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
	log.Infof("RPC Server [%s] listening on %s", e.name, e.server.GetAddr())

	done := make(chan struct{})
	go func() {
		e.started = true
		errChan <- e.server.Serve(e.ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		errChan <- nil
	}

	err = <-errChan
	if err != nil {
		log.Errorf("RPC Server [%s] start error: %s", e.name, err.Error())
		return err
	}

	if len(e.opts.startedHooks) > 0 {
		for _, fn := range e.opts.startedHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("RPC Server [%s] StartedHooks:%+v", e.name, err)
				return err
			}
		}
	}
	log.Infof("RPC Server [%s] start completed", e.name)
	return nil
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {
	err := e.server.Stop(ctx)
	if err != nil {
		log.Errorf("RPC Server [%s] stop error: %s", e.name, err.Error())
		return err
	}
	if len(e.opts.endHooks) > 0 {
		for _, fn := range e.opts.endHooks {
			err := fn(ctx)
			if err != nil {
				log.Errorf("RPC Server [%s] EndHook:", e.name, err)
				return err
			}
		}
	}
	log.Infof("RPC Server [%s] stop completed", e.name)
	return nil
}

// ServiceName 服务名称
func (s *Server) ServiceName() string {
	return s.opts.serviceName
}

func (e *Server) Endpoint() *url.URL {
	if e.endpoint == nil {
		e.endpoint = e.buildEndpoint()
	}
	return e.endpoint

}

func (e *Server) serverPath() string {
	return fmt.Sprintf("servers.%s", e.Name())
}

func (e *Server) buildEndpoint() *url.URL {
	host, port, err := xnet.ExtractHostPort(e.server.GetAddr())
	if err != nil {
		panic(fmt.Errorf("RPC Server Addr:%s 配置错误", e.server.GetAddr()))
	}
	if host == "" {
		host = global.LocalIp
	}
	return transport.NewEndpoint(e.server.GetProto(), fmt.Sprintf("%s:%d", host, port))
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
