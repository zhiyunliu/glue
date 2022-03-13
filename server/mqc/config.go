package mqc

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/velocity/server"
)

/*```
"mqc":{
			"config":{"addr":"redis://redisxxx","status":"start/stop"},
			"middlewares":[{},{}],
			"tasks":[{"queue":"xx.xx.xx","service":"/xx/bb/cc","status":"enable"},{"queue":"yy.yy.yy","service":"/xx/bb/yy"}],
		},
```*/
type Setting struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []server.Middleware `json:"middlewares"  yaml:"middlewares"`
	Tasks       server.Header       `json:"header"  yaml:"header"`
}

type Config struct {
	Addr              string        `json:"addr"`
	Status            server.Status `json:"status"`
	ReadTimeout       uint          `json:"read_timeout"`
	WriteTimeout      uint          `json:"write_timeout"`
	ReadHeaderTimeout uint          `json:"read_header_timeout"`
	MaxHeaderBytes    uint          `json:"max_header_bytes"`
}

type Task struct {
	Queue       string `json:"queue"`
	Service     string `json:"service,omitempty"`
	Enable      bool   `json:"enable"`
	Concurrency int    `json:"concurrency,omitempty"`
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
