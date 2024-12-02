package alloter

import (
	"sync"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/glue/engine"
)

var _ engine.AdapterEngine = (*AlloterEngine)(nil)

type AlloterEngine struct {
	Engine *alloter.Engine
	pool   sync.Pool
	opts   *engine.Options
}

func NewAlloterEngine(innerEngine *alloter.Engine, opts ...engine.Option) *AlloterEngine {
	g := &AlloterEngine{
		Engine: innerEngine,
		pool:   sync.Pool{},
		opts:   engine.DefaultOptions(),
	}
	alloter.SetMode(alloter.ReleaseMode)

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

		actx.Log().Errorf("No Method for %s,%s,clientip:%s", actx.Request().Path().FullPath(), actx.Request().GetMethod(), actx.Request().GetClientIP())

		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *AlloterEngine) NoRoute() {
	e.Engine.NoRoute(func(ctx *alloter.Context) {
		actx := e.pool.Get().(*AlloterContext)
		actx.reset(ctx)
		actx.opts = e.opts
		actx.Log().Errorf("[%s][%s]No Route for [%s]%s,clientip:%s", actx.ServerType(), actx.ServerName(), actx.Request().GetMethod(), actx.Request().Path().GetURL(), actx.Request().GetClientIP())
		actx.Close()
		e.pool.Put(actx)
	})
}

func (e *AlloterEngine) Handle(method string, path string, callfunc engine.HandlerFunc) {
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
func (e *AlloterEngine) HandleRequest(req engine.Request, resp engine.ResponseWriter) (err error) {
	return e.Engine.HandleRequest(req, resp)
}

func (e *AlloterEngine) GetImpl() any {
	return e.Engine
}
