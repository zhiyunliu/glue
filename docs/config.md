```json

	"app":{
		"encrypt":"1234567890123456",
		"mode":"debug",
		"ip_mask":"192.168",
		"dependencies":["After=network.target"],
		"options":{"LimitNOFILE":10240,"MaxOPENFiles":4096}
	},
	"registry":"nacos://aliyun",
 	"config":"nacos://aliyun",
	"dlocker":"redis://redis1",
	"caches":{
		"redisxxx":{"proto":"redis","addr":"redis://redis1"},
		"redisyyy":{"proto":"redis","addr":"redis://redis1"},
	},
	"queues":{
		"redisxxx":{"proto":"redis","addr":"redis://redis1"}
	},
	"rpcs":{
		"default":{"proto":"grpc","balancer":"round_robin","conn_timeout":10}
	},
	"redis":{
		"redis1":{"addrs":["192.168.0.1","192.168.0.2"],"auth":"","db":0,"dial_timeout":10,"read_timeout":10,"write_timeout":10,"pool_size":10}
	},
	"nacos":{
		"aliyun":{
			"encrypt":false,
			"client":{"namespace_id":"1cd02f66-fd24-4202-8009-32ffb0a3ac7e"},
			"server":[{"ipaddr":"192.168.0.120","port":8848}],
			"options":{"prefix":"api","group":"charge","cluster":"grey","weight":100}
		}
	},
	"prometheus":{
		"counter":{
			"namespace":"server",
			"subsystem":"requests",
			"name":"code_total",
			"help":"The total number of processed requests",
			"labels":["kind", "path", "code", "reason"]
		},
		"histogram":{
			"namespace":"server",
			"subsystem":"requests",
			"name":"duration_sec",
			"help":"server requests duration(sec)",
			"buckets":[0.05, 0.1, 0.5, 1, 1.5, 2, 2.5, 3, 4, 5],
			"labels":["kind", "path"]
		},
		"gauge":{
			"namespace":"server",
			"subsystem":"requests",
			"name":"cur_proc",
			"help":"server current processing.",
			"labels":["kind", "path"]
		}
	}
	"dbs":{
		"localhost":{"proto":"mysql","conn":"root:123456@tcp(localhost)/demo?charset=utf8","max_open":10,"max_idle":10,"life_time":100},
		"mssql":{"proto":"sqlserver","conn":"server=localohst;database=demos;uid=admin;pwd=123456;Min Pool Size=10;Max Pool Size=20","max_open":10,"max_idle":10,"life_time":100}
	},
	"servers":{
		"apiserver":{
			"config":{"addr":":8080","status":"start/stop","read_timeout":10,"write_timeout":10,"read_header_timeout":10,"max_header_bytes":65525},
			"header":{},
			"middlewares": [{"name": "metrics","data": {"proto": "prometheus"}}]
		},
		"rpcserver":{
			"config":{"addr":":8081","status":"start/stop","read_timeout":10,"connection_timeout":10,"read_buffer_size":32,"write_buffer_size":32, "max_recv_size":65535,"max_send_size":65535},
			"header":{},
			"middlewares": [{"name": "metrics","data": {"proto": "prometheus"}}]
		},
		"mqcserver":{
			"config":{"addr":"queues://redisxxx","status":"start/stop"},
			"middlewares": [{"name": "metrics","data": {"proto": "prometheus"}}],
			"tasks":[
				{"queue":"xx.xx.xx","service":"/xx/bb/cc","disable":true},
				{"queue":"yy.yy.yy","service":"/xx/bb/yy","concurrency":10,"disable_retry":true}
			],
		},
		"cronserver":{
			"config":{"status":"start/stop","sharding":1},
			"middlewares": [{"name": "metrics","data": {"proto": "prometheus"}}],
			"jobs":[
				{"cron":"* 15 2 * * ? *","service":"/xx/bb/cc","immediately":true,"monopoly":true,"disable":false},
				{"cron":"* 15 2 * * ? *","service":"/xx/bb/yy"}
			],
		}
	}
}

```

## IPC 的 Unix socket 共享目录

Linux/macOS 使用 `global.IPCSocketDir` 将 IPC 名称映射到统一绝对路径，默认目录为 `/tmp/glue-ipc`。例如服务 `debugbasic` 使用 `/tmp/glue-ipc/glue-debugbasic`，客户端和服务端不再依赖相同的工作目录。Windows 仍使用原来的命名管道，不读取此变量。

可以在编译时设置目录：

```bash
go build -ldflags "-X github.com/zhiyunliu/glue/global.IPCSocketDir=/run/glue-ipcsocket" .
```

也可以在应用启动前设置 `global.IPCSocketDir`，启动后不要修改。客户端与服务端必须使用相同目录，且该目录必须为绝对路径。

服务端自动创建目录，创建权限为 `0700`；已有目录必须为当前用户所有的真实目录，且不能允许组或其他用户写入。客户端只连接，不创建目录。默认仅支持同一操作系统用户运行的服务互通；使用其他用户首次创建的默认目录会启动失败，应为不同用户注入不同目录。使用 `/run` 时应由部署流程预先创建目录并设置所有者。

升级时需要同步升级并重启相关 Unix 服务端与客户端：旧版本使用工作目录下的 socket，新版本使用共享目录，两者不会自动兼容。不要删除运行中的 socket。应用层 `servers.ipcserver.config.addr` 仍填写服务名称，不填写目录；目录通过上述全局变量配置。
