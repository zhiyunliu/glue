package grpc

import (
	"encoding/json"
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/xrpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

const Proto = "grpc"

type clientResolver struct {
}

func (s *clientResolver) Name() string {
	return Proto
}

func (s *clientResolver) Resolve(name string, cfg config.Config) (xrpc.Client, error) {
	setval := &clientConfig{
		Name:        name,
		Balancer:    roundrobin.Name,
		ConnTimeout: 10,
	}
	err := cfg.ScanTo(setval)
	if err != nil {
		return nil, fmt.Errorf("读取grpc配置:%w", err)
	}
	if setval.Balancer == "" && len(setval.ServerConfig) == 0 {
		return nil, fmt.Errorf("grpc配置错误,必须指定Balancer/ServerConfig")
	}

	if len(setval.ServerConfig) == 0 {
		setval.ServerConfig = json.RawMessage(fmt.Sprintf(`{"LoadBalancingPolicy":"%s"}`, setval.Balancer))
	}
	return NewRequest(setval), nil
}

func init() {
	xrpc.RegisterClient(&clientResolver{})
}
