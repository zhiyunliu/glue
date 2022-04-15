package rpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/server"
	"github.com/zhiyunliu/gel/transport"
	"github.com/zhiyunliu/gel/xgrpc/grpcproto"
	"github.com/zhiyunliu/golibs/host"
	"google.golang.org/grpc"
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
	e.Options(WithConfig(cfg))
	cfg.Get(fmt.Sprintf("servers.%s", e.Name())).Scan(e.opts.setting)
}

// Start 开始
func (e *Server) Start(ctx context.Context) error {
	e.ctx = ctx
	if len(e.opts.startedHooks) > 0 {
		for _, fn := range e.opts.startedHooks {
			err := fn(ctx)
			if err != nil {
				log.Error("mqc.StartedHooks:", err)
				return err
			}
		}
	}

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

	log.Infof("RPC Server [%s] listening on %s", e.name, strings.ReplaceAll(lsr.Addr().String(), "[::]", global.LocalIp))

	go func() {

		e.started = true
		if err := e.srv.Serve(lsr); err != nil {
			log.Errorf("[%s] Server start error: %s", e.name, err.Error())
		}
		<-ctx.Done()

		if len(e.opts.endHooks) > 0 {
			for _, fn := range e.opts.endHooks {
				err := fn(ctx)
				if err != nil {
					log.Error("mqc.endHooks:", err)
				}
			}
		}
		err = e.Stop(ctx)
		if err != nil {
			log.Errorf("[%s] Server shutdown error: %s", e.name, err.Error())
		}
	}()

	return nil
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {
	e.srv.GracefulStop()
	return nil
}

func (e *Server) Endpoint() *url.URL {
	if e.endpoint == nil {
		e.endpoint = e.buildEndpoint()
	}
	return e.endpoint

}

func (e *Server) buildEndpoint() *url.URL {
	addr, err := host.Extract(e.opts.setting.Config.Addr)
	if err != nil {
		panic(fmt.Errorf("RPC Server Addr:%s 配置错误", e.opts.setting.Config.Addr))
	}
	return transport.NewEndpoint("grpc", addr)
}

func (e *Server) newProcessor() {
	var err error
	e.processor, err = newProcessor()
	if err != nil {
		panic(err)
	}

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
	e.opts.router.Handle(path, obj, server.MethodGet)
}
