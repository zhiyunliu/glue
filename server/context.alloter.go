package server

import (
	"context"

	vctx "github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/contrib/alloter"
	"github.com/zhiyunliu/velocity/log"
)

type AlloterContext struct {
	srvType string
	Actx    *alloter.Context

	areq   *ginRequest
	aresp  *gresponse
	logger log.Logger
}

func (ctx *AlloterContext) reset(gctx *alloter.Context) {
	ctx.Actx = gctx
	ctx.areq = nil
}

func (ctx *AlloterContext) ServerType() string {
	return ctx.srvType
}
func (ctx *AlloterContext) Context() context.Context {
	return ctx.Actx.Request.Context()
}

func (ctx *AlloterContext) ResetContext(nctx context.Context) {
	req := ctx.Actx.Request.WithContext(nctx)
	ctx.Actx.Request = req
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
func (ctx *AlloterContext) GetImpl() interface{} {
	return ctx.Actx
}
