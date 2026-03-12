package grpc

import (
	"encoding/json"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xrpc"
)

type clientConfig struct {
	Name         string          `json:"-"`
	ConnTimeout  int             `json:"conn_timeout"`
	Balancer     string          `json:"balancer"`      //负载类型 round_robin:论寻负载
	ServerConfig json.RawMessage `json:"server_config"` //
	Config       config.Config   `json:"-"`
}

type serverConfig struct {
	Config      *xrpc.Config        `json:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
}
