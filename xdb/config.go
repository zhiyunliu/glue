package xdb

import (
	"fmt"

	"github.com/zhiyunliu/golibs/xtypes"
)

type Config struct {
	Proto         string      `json:"proto" valid:"required"`
	Conn          string      `json:"conn" valid:"required" label:"连接字符串"`
	MaxOpen       int         `json:"max_open" valid:"required" label:"最大打开连接数"`
	MaxIdle       int         `json:"max_idle" valid:"required" label:"最大空闲连接数"`
	LifeTime      int         `json:"life_time" valid:"required" label:"单个连接时长(秒)"`
	ShowQueryLog  bool        `json:"show_query_log"  label:"开启慢查询日志"`
	LongQueryTime int         `json:"long_query_time"  label:"慢查询阈值(毫秒)"`
	LoggerName    string      `json:"logger_name" label:"日志提供程序"`
	Debug         bool        `json:"debug" label:"调试模式"`
	changefields  xtypes.XMap `json:"-"`
}

func (c *Config) saveChangeField(field string, val any) {
	if c.changefields == nil {
		c.changefields = make(xtypes.XMap)
	}
	c.changefields[field] = val
}

func (c Config) getUniqueKey() string {
	keys := c.changefields.Keys()
	val := make([]any, c.changefields.Len()*2)
	for i := range keys {
		val[i] = keys[i]
		val[i+1] = c.changefields[keys[i]]
	}
	return fmt.Sprint(val...)
}

var Default = &Config{
	MaxOpen:       10,
	MaxIdle:       5,
	LifeTime:      600,
	ShowQueryLog:  false,
	LongQueryTime: 500,
	LoggerName:    "dbslowsql",
	Debug:         false,
}

// Option 配置选项
type Option func(*Config)

func WithMaxOpen(maxOpen int) Option {
	return func(a *Config) {
		a.MaxOpen = maxOpen
		a.saveChangeField("MaxOpen", maxOpen)
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(a *Config) {
		a.MaxIdle = maxIdle
		a.saveChangeField("MaxIdle", maxIdle)

	}
}

func WithLongQueryTime(longQueryTime int) Option {
	return func(a *Config) {
		if longQueryTime <= 0 {
			return
		}
		a.LongQueryTime = longQueryTime
		a.saveChangeField("LongQueryTime", longQueryTime)

	}
}

func WithLifeTime(lifeTime int) Option {
	return func(a *Config) {
		if lifeTime <= 0 {
			return
		}
		a.LifeTime = lifeTime
		a.saveChangeField("LifeTime", lifeTime)

	}
}

func WithShowQueryLog(showQueryLog bool) Option {
	return func(a *Config) {
		a.ShowQueryLog = showQueryLog
		a.saveChangeField("ShowQueryLog", showQueryLog)

	}
}

func WithLoggerName(name string) Option {
	return func(a *Config) {
		a.LoggerName = name
		a.saveChangeField("LoggerName", name)
	}
}
