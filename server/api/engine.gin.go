package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.opts.handler.(*gin.Engine)
	server.RegistryEngineRoute(&server.GinEngine{
		Engine: engine,
		ERF:    e.opts.enc,
	}, e.opts.router)
}
