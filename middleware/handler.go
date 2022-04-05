package middleware

import "github.com/zhiyunliu/gel/context"

type Handler func(context.Context) interface{}

func (hfunc Handler) Handle(ctx context.Context) interface{} {
	return hfunc(ctx)
}
