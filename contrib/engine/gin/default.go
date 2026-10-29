package gin

import (
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (e *GinEngine) defaultHandle() {
	//healthcheck
	e.Engine.Handle(http.MethodGet, "/healthcheck", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	e.Engine.Handle(http.MethodGet, "/healthz", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	//metrics
	promHandler := promhttp.Handler()
	e.Engine.Handle(http.MethodGet, "/metrics", func(ctx *gin.Context) {
		promHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	//pprof
	pprof.Register(e.Engine)
}
