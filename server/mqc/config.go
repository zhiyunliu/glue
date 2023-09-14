package mqc

import (
	"github.com/zhiyunliu/glue/xmqc"
)

/*```
"mqc":{
			"config":{"addr":"redis://redisxxx","status":"start/stop"},
			"middlewares":[{},{}],
			"tasks":[{"queue":"xx.xx.xx","service":"/xx/bb/cc","disable":true},{"queue":"yy.yy.yy","service":"/xx/bb/yy"}],
		},
```*/

const Type string = "mqc"

type serverConfig struct {
	Config xmqc.Config `json:"config" yaml:"config"`
}
