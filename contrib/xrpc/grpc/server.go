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

func newServer(cfg *serverConfig,
	router *engine.RouterGroup,
	opts ...engine.Option) (server *Server, err error) {

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
		cfg:    cfg,
		srv:    grpc.NewServer(grpcOpts...),
		engine: alloter.New(),
	}

	server.processor, err = newProcessor(server.engine)
	if err != nil {
		return
	}
	adapterEngine := enginealloter.NewAlloterEngine(server.engine, opts...)
	engine.RegistryEngineRoute(adapterEngine, router)
	return
}

func (e *Server) GetAddr() string {
	return e.cfg.Addr
}

func (e *Server) GetProto() string {
	return Proto
}

func (e *Server) Serve(ctx context.Context) (err error) {
	newAddr, err := xnet.GetAvaliableAddr(log.DefaultLogger, global.LocalIp, e.cfg.Addr)
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
	return e.srv.Serve(lsr)
}

func (e *Server) Stop(ctx context.Context) error {
	e.srv.GracefulStop()
	return nil
}
