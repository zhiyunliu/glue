package defaultipc

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xipc"
)

const (
	Proto = "default"
)

var (
	_ xipc.ClientResolver = (*clientResolver)(nil)
	_ xipc.ServerResolver = (*serverResolver)(nil)
)

type clientResolver struct{}

func (clientResolver) Name() string { return Proto }
func (clientResolver) Resolve(proto string, cfg config.Config) (Client, error) {
	return NewClient(cfg)
}

type serverResolver struct{}

func (serverResolver) Name() string { return Proto }
func (serverResolver) Resolve(proto string, router *engine.RouterGroup, cfg config.Config, opts ...engine.Option) (Server, error) {
	setval := &serverConfig{
		Config: &xipc.Config{},
	}
	err := cfg.ScanTo(setval)
	if err != nil {
		return nil, fmt.Errorf("读取grpc配置[%s]错误%w", cfg.Path(), err)
	}

	return newServer(setval, router, opts...)
}

func init() {
	xipc.Register(clientResolver{})
	xipc.RegisterServer(serverResolver{})
}
