package log

import (
	"context"

	"github.com/zhiyunliu/golibs/xlog"
)

const _DEFAULT_BUILDER = "default"

type Builder interface {
	Name() string
	Build(context.Context, ...Option) Logger
	Close()
}

type defaultBuilderWrap struct {
}

func (b *defaultBuilderWrap) Name() string {
	return _DEFAULT_BUILDER
}

func (b *defaultBuilderWrap) Build(ctx context.Context, opts ...Option) Logger {
	return &Wraper{
		Logger: xlog.GetLogger(opts...),
	}
}

func (b *defaultBuilderWrap) Close() {
	xlog.Close()
}
