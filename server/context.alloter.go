package server

import (
	"context"

	vctx "github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/contrib/alloter"
	"github.com/zhiyunliu/velocity/log"
)

type AlloterContext struct {
	AloterCtx *alloter.Context
}

func (ctx *AlloterContext) Context() context.Context {
	return ctx.AloterCtx.Request.Context()
}

func (ctx *AlloterContext) ResetContext(nctx context.Context) {
	req := ctx.AloterCtx.Request.WithContext(nctx)
	ctx.AloterCtx.Request = req
}

func (ctx *AlloterContext) Header(key string) string {
	return ""
}

func (ctx *AlloterContext) Request() vctx.Request {
	return nil
}
func (ctx *AlloterContext) Response() vctx.Response {
	return nil
}
func (ctx *AlloterContext) Log() log.Logger {
	return nil
}
func (ctx *AlloterContext) Close() {

}
