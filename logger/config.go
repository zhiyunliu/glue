package logger

import (
	"sync"
)

var onceLock sync.Once
var config *ConfigSetting

func Config(opts ...Option) {
	onceLock.Do(func() {
		config = &ConfigSetting{}
		for i := range opts {
			opts[i]()
		}

	})
}

type ConfigSetting struct {
	DefaultData map[string]interface{}
	TimeStamp   string
}

type Option func()

func WithDefaultData(kvs map[string]interface{}) Option {
	return func() {
		config.DefaultData = kvs
	}
}
