package server

import (
	"sync"

	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/contrib/alloter"
	"github.com/zhiyunliu/velocity/global"
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
		return &GinContext{
			opts: g.opts,
		}
	}
	return g
}

func (e *AlloterEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(ctx *alloter.Context) {
		actx := e.pool.Get().(*AlloterContext)
		actx.reset(ctx)
		actx.srvType = e.opts.SrvType
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
