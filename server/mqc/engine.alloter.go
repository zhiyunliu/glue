package mqc

import (
	"github.com/zhiyunliu/velocity/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.processor.engine

	server.RegistryEngineRoute(&server.AlloterEngine{
		Engine: engine,
		ERF:    e.opts.enc,
	}, e.opts.router)

}
