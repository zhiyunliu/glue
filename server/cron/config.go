package cron

import (
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xcron"
)

/*```
"cron":{
			"config":{"status":"start/stop","sharding":1},
			"middlewares":[{},{}],
			"jobs":[{"cron":"* 15 2 * * ? *","service":"/xx/bb/cc","disable":false},{"cron":"* 15 2 * * ? *","service":"/xx/bb/yy"}],
		}
```*/

const Type string = "cron"

type serverConfig struct {
	Config      xcron.Config        `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
}
