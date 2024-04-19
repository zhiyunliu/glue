package queue

import (
	"fmt"
	"sort"

	"github.com/zhiyunliu/golibs/xtypes"
)

var (
	MaxRetrtCount int64 = 3 //最大重试三次
)

const (
	QueueKey string = "qk"
)

// Option 配置选项
type Option func(*Options)

type Options struct {
	CfgData map[string]any
}

func DefaultOptions() *Options {
	return &Options{
		CfgData: make(map[string]any),
	}
}

func (c Options) getUniqueKey() string {
	var changefields xtypes.XMap = c.CfgData
	keys := changefields.Keys()

	sort.Strings(keys)

	val := make([]any, changefields.Len()*2)
	for i := range keys {
		val[i] = keys[i]
		val[i+1] = changefields[keys[i]]
	}
	return fmt.Sprint(val...)
}

func (o *Options) setVal(k string, v any) {
	if len(o.CfgData) == 0 {
		o.CfgData = map[string]any{}
	}
	o.CfgData[k] = v
}

func WithOption(key string, val any) Option {
	return func(m *Options) {
		m.setVal(key, val)
	}
}

// func WithDBIndex(idx int) Option {
// 	return func(m *Options) {
// 		m.setVal("db", idx)
// 	}
// }

// func WithPoolSize(size int) Option {
// 	return func(m *Options) {
// 		m.setVal("pool_size", size)
// 	}
// }
