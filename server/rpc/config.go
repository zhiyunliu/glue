package rpc

import (
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/server"
)

/*```
	"rpc":{
			"config":{"addr":":8081","status":"start/stop","read_timeout":10,"connection_timeout":10,"read_buffer_size":32,"write_buffer_size":32, "max_recv_size":65535,"max_send_size":65535},
			"middlewares":[{},{}],
			"header":{},
		},
```*/

const Type string = "rpc"

type Setting struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
}

type Config struct {
	Addr           string        `json:"addr"`
	Status         server.Status `json:"status"`
	MaxRecvMsgSize int           `json:"max_recv_msg_size"`
	MaxSendMsgSize int           `json:"max_send_msg_size"`
}
