package server

import (
	"sync"

	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/contrib/alloter"
)

type AlloterEngine struct {
	Engine  *alloter.Engine
	ERF     EncodeResponseFunc
	pool    sync.Pool
	srtType string
}

func NewAlloterEngine(engine *alloter.Engine, erf EncodeResponseFunc) *AlloterEngine {
	g := &AlloterEngine{
		Engine: engine,
		ERF:    erf,
	}

	g.pool.New = func() interface{} {
		return &GinContext{
			erf: erf,
		}
	}
	return g
}

func (e *AlloterEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(ctx *alloter.Context) {
		actx := e.pool.Get().(*AlloterContext)
		actx.reset(ctx)
		actx.srvType = e.srtType
		callfunc(actx)
		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *AlloterEngine) EncodeResponseFunc(ctx context.Context, resp interface{}) error {
	return e.ERF(ctx, resp)
}
