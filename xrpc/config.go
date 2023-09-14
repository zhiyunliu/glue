package xrpc

import "github.com/zhiyunliu/glue/engine"

type Config struct {
	Addr           string        `json:"addr"`
	Proto          string        `json:"proto"`
	Status         engine.Status `json:"status"`
	MaxRecvMsgSize int           `json:"max_recv_msg_size"`
	MaxSendMsgSize int           `json:"max_send_msg_size"`
}
