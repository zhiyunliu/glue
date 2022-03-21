package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.opts.handler.(*gin.Engine)
	adapterEngine := server.NewGinEngine(engine, e.opts.enc)

	server.RegistryEngineRoute(adapterEngine, e.opts.router)
}
