package server

import (
	"context"

	"github.com/zhiyunliu/velocity/log"

	"github.com/gin-gonic/gin"
	vctx "github.com/zhiyunliu/velocity/context"
)

type GinContext struct {
	Gctx *gin.Context
}

func (ctx *GinContext) Context() context.Context {
	return ctx.Gctx.Request.Context()
}

func (ctx *GinContext) ResetContext(nctx context.Context) {
	req := ctx.Gctx.Request.WithContext(nctx)
	ctx.Gctx.Request = req
}

func (ctx *GinContext) Header(key string) string {
	return ""
}

func (ctx *GinContext) Request() vctx.Request {
	return nil
}
func (ctx *GinContext) Response() vctx.Response {
	return nil
}
func (ctx *GinContext) Log() log.Logger {
	return nil
}
func (ctx *GinContext) Close() {

}
