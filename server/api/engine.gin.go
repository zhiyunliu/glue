package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.opts.handler.(*gin.Engine)
	adapterEngine := server.NewGinEngine(engine,
		server.WithSrvType(e.Type()),
		server.WithErrorEncoder(e.opts.encErr),
		server.WithRequestDecoder(e.opts.decReq),
		server.WithResponseEncoder(e.opts.encResp))

	server.RegistryEngineRoute(adapterEngine, e.opts.router)
}
