package mqc

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/velocity/metadata"
	"github.com/zhiyunliu/velocity/server"
)

/*```
"mqc":{
			"config":{"addr":"redis://redisxxx","status":"start/stop"},
			"middlewares":[{},{}],
			"tasks":[{"queue":"xx.xx.xx","service":"/xx/bb/cc","disable":true},{"queue":"yy.yy.yy","service":"/xx/bb/yy"}],
		},
```*/

const Type string = "mqc"

type Setting struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []server.Middleware `json:"middlewares"  yaml:"middlewares"`
	Tasks       []*Task             `json:"tasks"  yaml:"tasks"`
}

type Config struct {
	Addr   string        `json:"addr"`
	Status server.Status `json:"status"`
}

type Task struct {
	Queue       string            `json:"queue"`
	Service     string            `json:"service,omitempty"`
	Disable     bool              `json:"disable"`
	Concurrency int               `json:"concurrency,omitempty"`
	Meta        metadata.Metadata `json:"meta,omitempty"`
}

func (t *Task) GetService() string {
	if t.Service != "" {
		return t.Service
	}
	tmp := t.Queue
	tmp = strings.ReplaceAll(tmp, ":", "_")
	tmp = strings.ReplaceAll(tmp, "/", "_")

	t.Service = fmt.Sprintf("/mqc_%s", tmp)
	return t.Service
}
