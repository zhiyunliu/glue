package server

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/transport"
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

// c := engine.pool.Get().(*Context)
// c.reset(r)
// c.writermem.reset()
// engine.handleRequest(c)
// if len(c.Errors) > 0 {
// 	err = c.Errors[0]
// }
// w = c.writermem.Copy()
// c.writermem.reset()
// c.reset(nil)
// engine.pool.Put(c)

func (e *GinEngine) Handle(method string, path string, callfunc HandlerFunc) {
	e.Engine.Handle(method, path, func(ctx *gin.Context) {
		actx := e.pool.Get().(*GinContext)
		actx.reset(ctx)

		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if s.timeout > 0 {
			ctx, cancel = context.WithTimeout(req.Context(), s.timeout)
		} else {
			ctx, cancel = context.WithCancel(req.Context())
		}
		defer cancel()

		pathTemplate := ctx.Request.URL.Path
		if route := mux.CurrentRoute(req); route != nil {
			// /path/123 -> /path/{id}
			pathTemplate, _ = route.GetPathTemplate()
		}
		tr := &Transport{
			endpoint:     s.endpoint.String(),
			operation:    pathTemplate,
			reqHeader:    headerCarrier(req.Header),
			replyHeader:  headerCarrier(w.Header()),
			request:      req,
			pathTemplate: pathTemplate,
		}
		ctx = transport.NewServerContext(ctx, tr)

		callfunc(actx)
		actx.Close()
		e.pool.Put(actx)
	})
}
func (e *GinEngine) EncodeResponseFunc(ctx context.Context, resp interface{}) error {
	return e.ERF(ctx, resp)
}
