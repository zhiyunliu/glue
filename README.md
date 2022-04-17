# gel
```text
清除git中的二进制等文件
git filter-branch --force --index-filter 'git rm --cached --ignore-unmatch examples/apiserver/apiserver' --prune-empty --tag-name-filter cat -- --all

```

``` json 总体配置结构
{
	"app":{
		"encrypt":"1234567890123456",
		"mode":"debug",
		"ip_mask":"192.168",
        "graceful_shutdown_timeout":15,
		"dependencies":["After=network.target"],
		"options":{"LimitNOFILE":10240,"MaxOPENFiles":4096}
	},
	"registry":"nacos://aliyun",
 	"config":"nacos://aliyun",
	"caches":{
		"redisxxx":"redis://redis1",
		"redisyyy":"redis://redis1",
	},
	"queues":{
		"redisxxx":"redis://redis1"
	},
	"rpcs":{
		"default":{"balancer":"round_robin","connection_timeout":10}
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
	"dbs":{
		"localhost":{"proto":"mysql","conn":"root:123456@tcp(localhost)/demo?charset=utf8","max_open":10,"max_idle":10,"life_time":100},
		"mssql":{"proto":"sqlserver","conn":"server=localohst;database=demos;uid=admin;pwd=123456;Min Pool Size=10;Max Pool Size=20","max_open":10,"max_idle":10,"life_time":100}
	},
	"servers":{
		"apiserver":{
			"config":{"addr":":8080","status":"start/stop","read_timeout":10,"write_timeout":10,"read_header_timeout":10,"max_header_bytes":65525},
			"middlewares":[{
				"auth":{"proto":"jwt","jwt":{},"exclude":["/**"]}
			}],			
			"header":{},
		},
		"rpcserver":{
			"config":{"addr":":8081","status":"start/stop","read_timeout":10,"connection_timeout":10,"read_buffer_size":32,"write_buffer_size":32, "max_recv_size":65535,"max_send_size":65535},
			"middlewares":[{},{}],
			"header":{},
		},
		"mqcserver":{
			"config":{"addr":"queues://redisxxx","status":"start/stop"},
			"middlewares":[{},{}],
			"tasks":[{"queue":"xx.xx.xx","service":"/xx/bb/cc","disable":true},{"queue":"yy.yy.yy","service":"/xx/bb/yy"}],
		},
		"cronserver":{
			"config":{"status":"start/stop","sharding":1},
			"middlewares":[{},{}],
			"jobs":[{"cron":"* 15 2 * * ? *","service":"/xx/bb/cc","disable":false},{"cron":"* 15 2 * * ? *","service":"/xx/bb/yy"}],
		}		
	}
}

```
```text 
1. 配置文件节点，只有registry 节点必须
2. "registry-1", "log-1", "db"  等都可以根据以“encrypt=”开头判定是否是加密串
3. 如果以加密方式存储，需要在环境变量中设置key,名称为当前应用程序的名称

```