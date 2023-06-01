package rpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	_ "github.com/zhiyunliu/glue/contrib/xrpc/grpc"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/golibs/xnet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	ctx       context.Context
	name      string
	processor *processor
	srv       *grpc.Server
	endpoint  *url.URL
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
	e.opts.setting.Config.Addr = server.GetAvaliableAddr(e.opts.setting.Config.Addr)
}

// Start 开始
func (e *Server) Start(ctx context.Context) error {
	if e.opts.setting.Config.Status == server.StatusStop {
		return nil
	}

	e.ctx = transport.WithServerContext(ctx, e)
	lsr, err := net.Listen("tcp", e.opts.setting.Config.Addr)
	if err != nil {
		return err
	}
	grpcOpts := []grpc.ServerOption{}
	if e.opts.setting.Config.MaxRecvMsgSize > 0 {
		grpcOpts = append(grpcOpts, grpc.MaxRecvMsgSize(e.opts.setting.Config.MaxRecvMsgSize))
	}
	if e.opts.setting.Config.MaxSendMsgSize > 0 {
		grpcOpts = append(grpcOpts, grpc.MaxSendMsgSize(e.opts.setting.Config.MaxSendMsgSize))
	}
	e.srv = grpc.NewServer(grpcOpts...)
	e.newProcessor()

	errChan := make(chan error, 1)
	log.Infof("RPC Server [%s] listening on %s%s", e.name, global.LocalIp, e.opts.setting.Config.Addr)
	go func() {
		e.started = true
		done := make(chan struct{})
		go func() {
			errChan <- e.srv.Serve(lsr)
			close(done)
		}()

		select {
		case <-done:
			return
		case <-time.After(time.Second):
			errChan <- nil
		}
	}()
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
	e.srv.GracefulStop()
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

func (e *Server) buildEndpoint() *url.URL {
	host, port, err := xnet.ExtractHostPort(e.opts.setting.Config.Addr)
	if err != nil {
		panic(fmt.Errorf("RPC Server Addr:%s 配置错误", e.opts.setting.Config.Addr))
	}
	if host == "" {
		host = global.LocalIp
	}
	return transport.NewEndpoint("grpc", fmt.Sprintf("%s:%d", host, port))
}

func (e *Server) newProcessor() {
	var err error
	e.processor, err = newProcessor(e)
	if err != nil {
		panic(err)
	}
	reflection.Register(e.srv)
	grpcproto.RegisterGRPCServer(e.srv, e.processor)

	e.registryEngineRoute()
}

func (e *Server) Use(middlewares ...middleware.Middleware) {
	e.opts.router.Use(middlewares...)
}

func (e *Server) Group(group string, middlewares ...middleware.Middleware) *server.RouterGroup {
	return e.opts.router.Group(group, middlewares...)
}

func (e *Server) Handle(path string, obj interface{}) {
	e.opts.router.Handle(path, obj, server.MethodPost)
}
