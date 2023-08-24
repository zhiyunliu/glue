package xdb

import (
	"reflect"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/xdb"
)

// DB 数据库配置
type Setting struct {
	Cfg           *xdb.Config   `json:"-"`
	slowThreshold time.Duration `json:"-"`
	logger        xdb.Logger    `json:"-"`
}

// New 构建DB连接信息
func NewConfig(opts ...xdb.Option) *Setting {
	db := &Setting{
		Cfg: &xdb.Config{
			MaxOpen:       xdb.Default.MaxOpen,
			MaxIdle:       xdb.Default.MaxIdle,
			LifeTime:      xdb.Default.LifeTime,
			ShowQueryLog:  xdb.Default.ShowQueryLog,
			LongQueryTime: xdb.Default.LongQueryTime,
			LoggerName:    xdb.Default.LoggerName,
			Debug:         xdb.Default.Debug,
		},
	}
	for _, opt := range opts {
		opt(db.Cfg)
	}
	return db
}

// 注册自定义类型转换
func RegisterDbType(dbType string, reflectType reflect.Type) error {
	return internal.RegisterDbType(dbType, reflectType)
}
