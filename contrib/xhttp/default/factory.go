package grpc

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/xhttp"
)

const Proto = "http"

type httpResolver struct {
}

func (s *httpResolver) Name() string {
	return Proto
}

func (s *httpResolver) Resolve(name string, cfg config.Config) (xhttp.Client, error) {
	setval := &setting{
		Name:                name,
		Balancer:            "random",
		KeepaliveTimeout:    15,
		ConnTimeout:         10,
		MaxIdleConns:        100,
		IdleConnTimeout:     90,
		TLSHandshakeTimeout: 10,
	}
	err := cfg.Scan(setval)
	if err != nil {
		return nil, fmt.Errorf("读取http配置:%w", err)
	}
	return NewRequest(setval), nil
}

func init() {
	xhttp.Register(&httpResolver{})
}
