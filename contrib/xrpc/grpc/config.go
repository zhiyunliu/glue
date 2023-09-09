package grpc

import (
	"encoding/json"

	"github.com/zhiyunliu/glue/config"
)

type clientConfig struct {
	Name         string          `json:"-"`
	ConnTimeout  int             `json:"conn_timeout"`
	Balancer     string          `json:"balancer"`      //负载类型 round_robin:论寻负载
	ServerConfig json.RawMessage `json:"server_config"` //
	Trace        bool            `json:"trace"`
	Config       config.Config   `json:"-"`
}

type serverConfig struct {
	Addr           string `json:"addr"`
	MaxRecvMsgSize int    `json:"max_recv_msg_size"`
	MaxSendMsgSize int    `json:"max_send_msg_size"`
}
