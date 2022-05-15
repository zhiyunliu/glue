package mqc

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/gel/metadata"
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/server"
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
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
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
	t.Service = getService(t.Queue)
	return t.Service
}

func getService(queue string) string {
	if strings.HasPrefix(queue, "/") {
		return queue
	}
	tmp := queue
	tmp = strings.ReplaceAll(tmp, ":", "_")
	return fmt.Sprintf("/mqc_%s", tmp)
}
