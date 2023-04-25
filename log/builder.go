package log

import (
	"github.com/zhiyunliu/golibs/xlog"
)

const _DEFAULT_BUILDER = "default"

type Builder interface {
	Name() string
	Build(...Option) Logger
	Close()
}

type defaultBuilderWrap struct {
}

func (b *defaultBuilderWrap) Name() string {
	return _DEFAULT_BUILDER
}

func (b *defaultBuilderWrap) Build(opts ...Option) Logger {

	return &wraper{
		xloger: xlog.New(opts...),
	}
}

func (b *defaultBuilderWrap) Close() {
	xlog.Close()
}
