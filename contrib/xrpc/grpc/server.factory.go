package grpc

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xrpc"
)

type serverResolver struct {
}

func (s *serverResolver) Name() string {
	return Proto
}

func (s *serverResolver) Resolve(name string,
	router *engine.RouterGroup,
	cfg config.Config,
	opts ...engine.Option) (xrpc.Server, error) {
	setval := &serverConfig{
		Config: &xrpc.Config{},
	}
	err := cfg.ScanTo(setval)
	if err != nil {
		return nil, fmt.Errorf("读取grpc配置[%s]错误%w", cfg.Path(), err)
	}

	return newServer(setval, router, opts...)
}

func init() {
	xrpc.RegisterServer(&serverResolver{})
}
