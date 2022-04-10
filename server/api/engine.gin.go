package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/gel/server"
)

func (e *Server) registryEngineRoute() {
	gin.SetMode("release")
	engine := gin.New()
	e.opts.handler = engine
	adapterEngine := server.NewGinEngine(engine,
		server.WithSrvType(e.Type()),
		server.WithErrorEncoder(e.opts.encErr),
		server.WithRequestDecoder(e.opts.decReq),
		server.WithResponseEncoder(e.opts.encResp))

	server.RegistryEngineRoute(adapterEngine, e.opts.router)
}
