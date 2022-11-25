package xdb

import (
	"time"

	"github.com/zhiyunliu/glue/xdb"
)

//DB 数据库配置
type Config struct {
	Proto         string        `json:"proto" valid:"required"`
	Conn          string        `json:"conn" valid:"required" label:"连接字符串"`
	MaxOpen       int           `json:"max_open" valid:"required" label:"最大打开连接数"`
	MaxIdle       int           `json:"max_idle" valid:"required" label:"最大空闲连接数"`
	LifeTime      int           `json:"life_time" valid:"required" label:"单个连接时长(秒)"`
	ShowQueryLog  bool          `json:"show_query_log"  label:"开启慢查询日志"`
	LongQueryTime int           `json:"long_query_time"  label:"慢查询阈值(毫秒)"`
	LoggerName    string        `json:"logger_name" label:"日志提供程序"`
	Debug         bool          `json:"debug" label:"调试模式"`
	slowThreshold time.Duration `json:"-"`
	logger        xdb.Logger    `json:"-"`
}

//New 构建DB连接信息
func New(proto string, conn string, opts ...Option) *Config {
	db := &Config{
		Proto:         proto,
		Conn:          conn,
		MaxOpen:       10,
		MaxIdle:       10,
		LifeTime:      600,
		LongQueryTime: 500,
	}
	for _, opt := range opts {
		opt(db)
	}
	return db
}

//Option 配置选项
type Option func(*Config)

func WithMaxOpen(maxOpen int) Option {
	return func(a *Config) {
		a.MaxOpen = maxOpen
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(a *Config) {
		a.MaxIdle = maxIdle
	}
}

func WithLongQueryTime(longQueryTime int) Option {
	return func(a *Config) {
		if longQueryTime <= 0 {
			return
		}
		a.LongQueryTime = longQueryTime
	}
}

func WithLifeTime(lifeTime int) Option {
	return func(a *Config) {
		a.LifeTime = lifeTime
	}
}

func WithShowQueryLog(showQueryLog bool) Option {
	return func(a *Config) {
		a.ShowQueryLog = showQueryLog
	}
}

func WithLoggerName(name string) Option {
	return func(a *Config) {
		a.LoggerName = name
	}
}
