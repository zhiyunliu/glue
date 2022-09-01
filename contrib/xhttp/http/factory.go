package http

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	_ "github.com/zhiyunliu/glue/selector/p2c"
	_ "github.com/zhiyunliu/glue/selector/random"
	_ "github.com/zhiyunliu/glue/selector/wrr"
	"github.com/zhiyunliu/glue/xhttp"
)

const Proto = "xhttp"

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
