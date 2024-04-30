package api

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

/*
```

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
	}

```
*/
type serverConfig struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
	Header      engine.Header       `json:"header"  yaml:"header"`
}

type Config struct {
	Addr              string        `json:"addr"`
	Engine            string        `json:"engine"`
	Status            engine.Status `json:"status"`
	ReadTimeout       uint          `json:"read_timeout"`
	WriteTimeout      uint          `json:"write_timeout"`
	ReadHeaderTimeout uint          `json:"read_header_timeout"`
	MaxHeaderBytes    uint          `json:"max_header_bytes"`
}
