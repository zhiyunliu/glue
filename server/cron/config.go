package cron

import (
	"fmt"
	"sort"
	"strings"
	"time"

	cron "github.com/robfig/cron/v3"
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
	Cron                string            `json:"cron"`
	Service             string            `json:"service"`
	Disable             bool              `json:"disable"`
	Immediately         bool              `json:"immediately"`
	Monopoly            bool              `json:"monopoly"`
	Meta                metadata.Metadata `json:"meta,omitempty"`
	schedule            cron.Schedule     `json:"-"`
	immediatelyExecuted bool              `json:"-"`
}

func (t *Job) GetKey() string {
	mks := make([]string, 0)
	for k := range t.Meta {
		if t.Meta[k] == "" {
			continue
		}
		mks = append(mks, k)
	}
	sort.Strings(mks)
	for i := range mks {
		k := mks[i]
		mks[i] = fmt.Sprintf("k:%s,v:%s", k, t.Meta[k])
	}
	tmpKey := fmt.Sprintf("c:%s,s:%s,m:%s", t.Cron, t.Service, strings.Join(mks, ","))
	return md5.Str(tmpKey)
}

//服务地址
func (t *Job) GetService() string {
	return t.Service
}

//是否立即执行
func (t *Job) IsImmediately() bool {
	return t.Immediately
}

//是否独占
func (t *Job) IsMonopoly() bool {
	return t.Monopoly
}

//NextTime 下次执行时间
func (m *Job) NextTime(t time.Time) time.Time {
	if m.IsImmediately() && !m.immediatelyExecuted {
		m.immediatelyExecuted = true
		return t
	}
	return m.schedule.Next(t)
}
