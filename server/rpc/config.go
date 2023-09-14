package rpc

import "github.com/zhiyunliu/glue/xrpc"

/*```
	"rpc":{
			"config":{"addr":":8081","status":"start/stop","read_timeout":10,"connection_timeout":10,"read_buffer_size":32,"write_buffer_size":32, "max_recv_size":65535,"max_send_size":65535},
			"middlewares":[{},{}],
			"header":{},
		},
```*/

const Type string = "rpc"

type serverConfig struct {
	Config xrpc.Config `json:"config" yaml:"config"`
}
