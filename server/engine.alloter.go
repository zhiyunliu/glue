package server

import (
	"sync"

	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/contrib/alloter"
	"github.com/zhiyunliu/gel/global"
)

type AlloterEngine struct {
	Engine *alloter.Engine
	pool   sync.Pool
	opts   *options
}

func NewAlloterEngine(engine *alloter.Engine, opts ...Option) *AlloterEngine {
	g := &AlloterEngine{
		Engine: engine,
		opts:   setDefaultOptions(),
	}
	alloter.SetMode(global.Mode)

	for i := range opts {
		opts[i](g.opts)
	}
	g.pool.New = func() interface{} {
		return newAlloterContext(g.opts)
	}
	return g
}

func (e *AlloterEngine) NoMethod() {
	e.Engine.NoMethod(func(ctx *alloter.Context) {
		actx := e.pool.Get().(*AlloterContext)
		actx.reset(ctx)
		actx.opts = e.opts

		actx.Log().Errorf("No Method for %s", actx.Request().Path().FullPath())

		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *AlloterEngine) NoRoute() {
	e.Engine.NoRoute(func(ctx *alloter.Context) {
		actx := e.pool.Get().(*AlloterContext)
		actx.reset(ctx)
		actx.opts = e.opts
		actx.Log().Errorf("[%s][%s]No Route for [%s]%s", actx.ServerType(), actx.ServerName(), ctx.Request.GetMethod(), actx.Request().Path().GetURL())
		actx.Close()
		e.pool.Put(actx)
	})
}

func (e *AlloterEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(ctx *alloter.Context) {
		actx := e.pool.Get().(*AlloterContext)
		actx.reset(ctx)
		actx.opts = e.opts
		callfunc(actx)
		actx.Close()
		e.pool.Put(actx)
	})
}

func (e *AlloterEngine) Write(ctx context.Context, resp interface{}) {
	err := ctx.Response().Write(resp)
	if err != nil {
		ctx.Log().Errorf("%s:写入响应出错:%s,%+v", e.opts.SrvType, ctx.Request().Path().FullPath(), err)
	}
}
