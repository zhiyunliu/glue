package xcron

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	cron "github.com/robfig/cron/v3"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

type Config struct {
	Addr   string        `json:"addr"`
	Status engine.Status `json:"status"`
	Proto  string        `json:"proto"`
}

type Job struct {
	Cron                string            `json:"cron"`
	Service             string            `json:"service"`
	Disable             bool              `json:"disable"`
	Immediately         bool              `json:"immediately"`
	Monopoly            bool              `json:"monopoly"`
	WithSeconds         bool              `json:"with_seconds"`
	Meta                metadata.Metadata `json:"meta,omitempty"`
	schedule            cron.Schedule     `json:"-"`
	immediatelyExecuted bool              `json:"-"`
	tmpKey              string            `json:"-"`
	DlockKey            string            `json:"-"`
}

func (t *Job) GetKey() string {
	if t.tmpKey != "" {
		return t.tmpKey
	}
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
	orgKey := fmt.Sprintf("c:%s,s:%s,m:%s", t.Cron, t.Service, strings.Join(mks, ","))
	t.tmpKey = md5.Str(orgKey)
	return t.tmpKey
}

// 服务地址
func (t *Job) GetLockData() string {
	return fmt.Sprintf("%s(c:%s,s:%s,m:%+v)", global.LocalIp, t.Cron, t.Service, t.Meta.String())
}

// 服务地址
func (t *Job) GetService() string {
	return t.Service
}

// 是否立即执行
func (t *Job) IsImmediately() bool {
	return t.Immediately
}

// 是否独占
func (t *Job) IsMonopoly() bool {
	return t.Monopoly
}

// NextTime 下次执行时间
func (m *Job) NextTime(t time.Time) (nextTime time.Time) {
	if m.IsImmediately() && !m.immediatelyExecuted {
		m.immediatelyExecuted = true
		return t
	}
	nextTime = m.schedule.Next(t)
	return nextTime
}

func (m *Job) CalcExpireSeconds() int {
	nextTime := m.schedule.Next(time.Now())
	val := math.Ceil(time.Until(nextTime).Seconds())
	return int(val)
}

func (m *Job) Init() (err error) {
	if m.schedule != nil {
		return nil
	}
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	if m.WithSeconds {
		parser = cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)
	}
	m.schedule, err = parser.Parse(m.Cron)
	if err != nil {
		err = fmt.Errorf("cron parser.Parse:%s,err:%+v", m.Cron, err)
	}
	return err
}
