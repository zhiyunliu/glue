package log

import (
	"github.com/zhiyunliu/golibs/xlog"
)

type Builder interface {
	Build(...Option) Logger
	Close()
}

type defaultBuilderWrap struct {
}

func (b *defaultBuilderWrap) Build(opts ...Option) Logger {

	return &wraper{
		xloger: xlog.New(opts...),
	}
}

func (b *defaultBuilderWrap) Close() {
	xlog.Close()
}
