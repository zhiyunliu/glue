package server

import (
	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/contrib/alloter"
)

type AlloterEngine struct {
	Engine *alloter.Engine
	ERF    EncodeResponseFunc
}

func (e *AlloterEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(ctx *alloter.Context) {
		actx := &AlloterContext{
			AloterCtx: ctx,
		}
		callfunc(actx)
	})
}
func (e *AlloterEngine) EncodeResponseFunc(ctx context.Context, resp interface{}) error {
	return e.ERF(ctx, resp)
}
