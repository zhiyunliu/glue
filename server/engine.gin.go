package server

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/context"
)

type GinEngine struct {
	Engine *gin.Engine
	ERF    EncodeResponseFunc
	pool   sync.Pool
}

func NewGinEngine(engine *gin.Engine, erf EncodeResponseFunc) *GinEngine {
	g := &GinEngine{
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

func (e *GinEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(gctx *gin.Context) {
		actx := e.pool.Get().(*GinContext)
		actx.reset(gctx)
		actx.srvType = "api"
		callfunc(actx)
		actx.Gctx.Writer.Flush()
		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *GinEngine) EncodeResponseFunc(ctx context.Context, resp interface{}) error {
	return e.ERF(ctx, resp)
}
