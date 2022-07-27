package cron

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

/*```
"cron":{
			"config":{"status":"start/stop","sharding":1},
			"middlewares":[{},{}],
			"jobs":[{"cron":"* 15 2 * * ? *","service":"/xx/bb/cc","disable":false},{"cron":"* 15 2 * * ? *","service":"/xx/bb/yy"}],
		}
```*/

const Type string = "cron"

type Setting struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
	Jobs        []*Job              `json:"jobs"  yaml:"jobs"`
}

type Config struct {
	Status server.Status `json:"status"`
}

type Job struct {
	Cron        string            `json:"cron"`
	Service     string            `json:"service"`
	Disable     bool              `json:"disable"`
	Immediately bool              `json:"immediately"`
	Meta        metadata.Metadata `json:"meta,omitempty"`
}

func (t *Job) GetKey() string {

	mks := make([]string, len(t.Meta))
	i := 0
	for k := range t.Meta {
		mks[i] = k
	}
	sort.Strings(mks)
	for i := range mks {
		k := mks[i]
		if t.Meta[k] == "" {
			continue
		}
		mks[i] = fmt.Sprintf("k:%s,v:%s", k, t.Meta[k])
	}
	tmpKey := fmt.Sprintf("c:%s,s:%s,m:%s", t.Cron, t.Service, strings.Join(mks, ","))
	return md5.Str(tmpKey)
}

func (t *Job) GetService() string {
	return t.Service
}
func (t *Job) IsImmediately() bool {
	return t.Immediately
}
