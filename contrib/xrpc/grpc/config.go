package grpc

import "github.com/zhiyunliu/glue/config"

type setting struct {
	Name        string        `json:"-"`
	ConnTimeout int           `json:"conn_timeout"`
	Balancer    string        `json:"balancer"` //负载类型 round_robin:论寻负载
	Trace       bool          `json:"trace"`
	Config      config.Config `json:"-"`
}
