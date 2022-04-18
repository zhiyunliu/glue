package grpc

import "github.com/zhiyunliu/gel/config"

type setting struct {
	Name         string        `json:"-"`
	ConntTimeout int           `json:"connection_timeout"`
	Balancer     string        `json:"balancer"` //负载类型 round_robin:论寻负载
	Config       config.Config `json:"-"`
}
