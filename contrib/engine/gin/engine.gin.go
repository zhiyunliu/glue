package gin

import (
	"net/http"
	"sync"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
)

var _ engine.AdapterEngine = (*GinEngine)(nil)

type GinEngine struct {
	Engine *gin.Engine
	pool   sync.Pool
	opts   *engine.Options
}

func NewGinEngine(ginEngine *gin.Engine, opts ...engine.Option) engine.AdapterEngine {
	g := &GinEngine{
		Engine: ginEngine,
		opts:   engine.DefaultOptions(),
	}
	for i := range opts {
		opts[i](g.opts)
	}
	g.pool.New = func() interface{} {
		return newGinContext(g.opts)
	}
	g.defaultHandle()
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
		actx.Log().Errorf("[%s][%s]No Route for [%s]%s", actx.ServerType(), actx.ServerName(), ctx.Request.Method, actx.Request().Path().GetURL())
		actx.Close()
		e.pool.Put(actx)
	})
}

func (e *GinEngine) Handle(method string, path string, callfunc engine.HandlerFunc) {
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

func (e *GinEngine) HandleRequest(req engine.Request, resp engine.ResponseWriter) (err error) {
	return
}

func (e *GinEngine) GetImpl() any {
	return &httpEngine{engine: e.Engine}
}

func (e *GinEngine) defaultHandle() {
	e.Engine.Handle(http.MethodGet, "/healthcheck", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	pprof.Register(e.Engine)

	promHandler := promhttp.Handler()
	e.Engine.Handle(http.MethodGet, "/metrics", func(ctx *gin.Context) {
		promHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})
}

type httpEngine struct {
	engine *gin.Engine
}

func (e *httpEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.engine.ServeHTTP(w, req)
}
func (e *httpEngine) StaticFile(relativePath string, filepath string) {
	e.engine.StaticFile(relativePath, filepath)
}
func (e *httpEngine) Static(relativePath string, root string) {
	e.engine.Static(relativePath, root)
}
