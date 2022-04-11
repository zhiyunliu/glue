package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/server"
)

func (e *Server) registryEngineRoute() {
	gin.SetMode(global.Mode)
	engine := gin.New()
	e.opts.handler = engine
	adapterEngine := server.NewGinEngine(engine,
		server.WithSrvType(e.Type()),
		server.WithErrorEncoder(e.opts.encErr),
		server.WithRequestDecoder(e.opts.decReq),
		server.WithResponseEncoder(e.opts.encResp))

	engine.Handle(http.MethodGet, "/healthcheck", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "success")
	})

	server.RegistryEngineRoute(adapterEngine, e.opts.router)
}
