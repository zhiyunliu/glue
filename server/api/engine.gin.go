package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/middleware"

	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/server"
)

func (e *Server) registryEngineRoute() {
	gin.SetMode(global.Mode)
	engine := gin.New()
	e.opts.handler = engine
	adapterEngine := server.NewGinEngine(engine,
		server.WithSrvType(e.Type()),
		server.WithSrvName(e.Name()),
		server.WithErrorEncoder(e.opts.encErr),
		server.WithRequestDecoder(e.opts.decReq),

		server.WithResponseEncoder(func(ctx context.Context, resp interface{}) error {
			for k, v := range e.opts.setting.Header {
				ctx.Response().Header(k, v)
			}
			return e.opts.encResp(ctx, resp)
		}))

	engine.Handle(http.MethodGet, "/healthcheck", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	promHandler := promhttp.Handler()
	engine.Handle(http.MethodGet, "/metrics", func(ctx *gin.Context) {
		promHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	for _, m := range e.opts.setting.Middlewares {
		e.opts.router.Use(middleware.Resolve(&m))
	}

	for _, s := range e.opts.static {
		if s.IsFile {
			engine.StaticFile(s.RouterPath, s.FilePath)
		} else {
			engine.Static(s.RouterPath, s.FilePath)
		}
	}
	server.RegistryEngineRoute(adapterEngine, e.opts.router, e.opts.logOpts)
}
