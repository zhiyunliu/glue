package xcron

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	cron "github.com/robfig/cron/v3"
	"github.com/zhiyunliu/glue/engine"
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
func (m *Job) NextTime(t time.Time) time.Time {
	if m.IsImmediately() && !m.immediatelyExecuted {
		m.immediatelyExecuted = true
		return t
	}
	return m.schedule.Next(t)
}

func (m *Job) CalcExpireTime() int {
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
	return err
}
