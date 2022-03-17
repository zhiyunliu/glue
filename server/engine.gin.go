package server

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/context"
)

type GinEngine struct {
	Engine *gin.Engine
	ERF    EncodeResponseFunc
}

func (e *GinEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(ctx *gin.Context) {
		actx := &GinContext{
			Gctx: ctx,
		}
		callfunc(actx)
	})
}
func (e *GinEngine) EncodeResponseFunc(ctx context.Context, resp interface{}) error {
	return e.ERF(ctx, resp)
}
