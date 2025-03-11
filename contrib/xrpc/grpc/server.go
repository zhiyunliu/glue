package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/zhiyunliu/glue/contrib/alloter"
	enginealloter "github.com/zhiyunliu/glue/contrib/engine/alloter"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/golibs/xnet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	cfg       *serverConfig
	srv       *grpc.Server
	engine    *alloter.Engine
	processor *processor
}

func newServer(srvcfg *serverConfig,
	router *engine.RouterGroup,
	opts ...engine.Option) (server *Server, err error) {

	cfg := srvcfg.Config
	grpcOpts := []grpc.ServerOption{}
	if cfg.MaxRecvMsgSize > 0 {
		grpcOpts = append(grpcOpts, grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize))
	}
	if cfg.MaxSendMsgSize > 0 {
		grpcOpts = append(grpcOpts, grpc.MaxSendMsgSize(cfg.MaxSendMsgSize))
	}
	cfg.Addr, err = xnet.GetAvaliableAddr(log.DefaultLogger, global.LocalIp, cfg.Addr)
	if err != nil {
		err = fmt.Errorf("GRPC Avaliable Addr %+v", err)
		return
	}

	server = &Server{
		cfg:    srvcfg,
		srv:    grpc.NewServer(grpcOpts...),
		engine: alloter.New(),
	}

	server.processor, err = newProcessor(server.engine)
	if err != nil {
		return
	}

	midwares, err := middleware.BuildMiddlewareList(srvcfg.Middlewares)
	if err != nil {
		err = fmt.Errorf("engine:[%s] BuildMiddlewareList,%w", srvcfg.Config.Proto, err)
		return
	}
	router.Use(midwares...)

	adapterEngine := enginealloter.NewAlloterEngine(server.engine, opts...)
	engine.RegistryEngineRoute(adapterEngine, router)
	return
}

func (e *Server) GetAddr() string {
	return e.cfg.Config.Addr
}

func (e *Server) GetProto() string {
	return Proto
}

func (e *Server) Serve(ctx context.Context) (err error) {
	newAddr, err := xnet.GetAvaliableAddr(log.DefaultLogger, global.LocalIp, e.GetAddr())
	if err != nil {
		return
	}
	reflection.Register(e.srv)
	grpcproto.RegisterGRPCServer(e.srv, e.processor)

	if err != nil {
		return
	}
	lsr, err := net.Listen("tcp", newAddr)
	if err != nil {
		return err
	}

	err = e.srv.Serve(lsr)
	if err != nil && err != grpc.ErrServerStopped {
		return err
	}
	return nil
}

func (e *Server) Stop(ctx context.Context) error {
	if e.srv != nil {
		e.srv.GracefulStop()
	}
	return nil
}
