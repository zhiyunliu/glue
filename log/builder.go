package log

import (
	"github.com/zhiyunliu/golibs/xlog"
)

type Builder interface {
	Build(...Option) Logger
}

type defaultBuilder struct {
}

func (b *defaultBuilder) Build(opts ...Option) Logger {
	return &wraper{
		xloger: xlog.New(opts...),
	}
}
