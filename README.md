# gel

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
	"cache":{
		"redisxxx":"redis://redis1",
		"redisyyy":"redis://redis1",
	},
	"queue":{
		"redisxxx":"redis://redis1",
		"nsq11":{"proto":"nsq","xx":"xx"},
		"loacl22":{"proto":"loacl","xx":"xx"},
	},
	"redis":{
		"redis1":{"addrs":["192.168.0.1","192.168.0.2"],"auth":"","db":"","dial_timeout":10,"read_timeout":10,"write_timeout":10,"pool_size":10}
	},
	"nacos":{
		"aliyun":{
			"encrypt":"false",
			"client":{"NamespaceId":"1cd02f66-fd24-4202-8009-32ffb0a3ac7e"},
			"server":[{"IpAddr":"192.168.0.120","Port":8848}],
			"options":{"kind":"api","group":"charge","cluster":"grey"}
		}
	},
	"db":{
		"db1":{ "proto":"mysql","conn_str":"","max_open":100,"max_idle":100,"life_time":100},
		"db2":{ "proto":"ora","conn_str":"","max_open":100,"max_idle":100,"life_time":100},
	},
	"servers":{
		"api":{
			"config":{"addr":":8080","status":"start/stop","read_timeout":10,"write_timeout":10,"read_header_timeout":10,"max_header_bytes":65525},
			"middlewares":[
			{
				"auth":{
					"proto":"jwt",
					"jwt":{},
					"exclude":["/**"]
				}
			},{}],			
			"header":{},
		},
		"rpc":{
			"config":{"addr":":8081","status":"start/stop","read_timeout":10,"connection_timeout":10,"read_buffer_size":32,"write_buffer_size":32, "max_recv_size":65535,"max_send_size":65535},
			"middlewares":[{},{}],
			"header":{},
		},
		"mqc":{
			"config":{"addr":"redis://redisxxx","status":"start/stop"},
			"middlewares":[{},{}],
			"tasks":[{"queue":"xx.xx.xx","service":"/xx/bb/cc","disable":true},{"queue":"yy.yy.yy","service":"/xx/bb/yy"}],
		},
		"cron":{
			"config":{"status":"start/stop","sharding":1},
			"middlewares":[{},{}],
			"job":[{"cron":"* 15 2 * * ? *","service":"/xx/bb/cc","status":"enable"},{"cron":"* 15 2 * * ? *","service":"/xx/bb/yy"}],
		}		
	}
}

```
```text 
1. 配置文件节点，只有registry 节点必须
2. "registry-1", "log-1", "db"  等都可以根据以“encrypt=”开头判定是否是加密串
3. 如果以加密方式存储，需要在环境变量中设置key,名称为当前应用程序的名称

```