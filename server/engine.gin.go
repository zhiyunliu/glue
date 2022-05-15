package server

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/global"
)

type GinEngine struct {
	Engine *gin.Engine
	pool   sync.Pool
	opts   *options
}

func NewGinEngine(engine *gin.Engine, opts ...Option) *GinEngine {
	g := &GinEngine{
		Engine: engine,
		opts:   setDefaultOptions(),
	}
	gin.SetMode(global.Mode)
	for i := range opts {
		opts[i](g.opts)
	}
	g.pool.New = func() interface{} {
		return newGinContext(g.opts)
	}
	return g
}
func (e *GinEngine) NoMethod() {
	e.Engine.NoMethod(func(ctx *gin.Context) {
		actx := e.pool.Get().(*GinContext)
		actx.reset(ctx)
		actx.opts = e.opts

		actx.Log().Errorf("No Method for %s", actx.Request().Path().FullPath())

		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *GinEngine) NoRoute() {
	e.Engine.NoRoute(func(ctx *gin.Context) {
		actx := e.pool.Get().(*GinContext)
		actx.reset(ctx)
		actx.opts = e.opts
		actx.Log().Errorf("[%s][%s]No Route for [%s]%s", actx.ServerType(), actx.ServerName(), ctx.Request.Method, actx.Request().Path().FullPath())
		actx.Close()
		e.pool.Put(actx)
	})
}

func (e *GinEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(gctx *gin.Context) {
		actx := e.pool.Get().(*GinContext)
		actx.reset(gctx)
		actx.opts = e.opts
		callfunc(actx)
		actx.Gctx.Writer.Flush()
		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *GinEngine) Write(ctx context.Context, resp interface{}) {
	err := ctx.Response().Write(resp)
	if err != nil {
		ctx.Log().Errorf("%s:写入响应出错:%s,%+v", e.opts.SrvType, ctx.Request().Path().FullPath(), err)
	}
}
