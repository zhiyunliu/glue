package mqc

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
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
	Tasks       TaskList            `json:"tasks"  yaml:"tasks"`
}

type Config struct {
	Addr   string        `json:"addr"`
	Engine string        `json:"engine"`
	Status server.Status `json:"status"`
}

func (c Config) String() string {
	return c.Addr
}

type Task struct {
	Queue       string            `json:"queue"`
	Service     string            `json:"service,omitempty"`
	Disable     bool              `json:"disable"`
	Concurrency int               `json:"concurrency,omitempty"`
	Meta        metadata.Metadata `json:"meta,omitempty"`
}

type TaskList []*Task

func (t *Task) GetQueue() string {
	return t.Queue
}

func (t *Task) GetConcurrency() int {
	return t.Concurrency
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
