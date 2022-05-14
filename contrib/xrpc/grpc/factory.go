package grpc

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/xrpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

const Proto = "grpc"

type grpcResolver struct {
}

func (s *grpcResolver) Name() string {
	return Proto
}

func (s *grpcResolver) Resolve(name string, cfg config.Config) (xrpc.Client, error) {
	setval := &setting{
		Name:        name,
		Balancer:    roundrobin.Name,
		ConnTimeout: 10,
	}
	err := cfg.Scan(setval)
	if err != nil {
		return nil, fmt.Errorf("读取grpc配置:%w", err)
	}
	return NewRequest(setval), nil
}

func init() {
	xrpc.Register(&grpcResolver{})
}
