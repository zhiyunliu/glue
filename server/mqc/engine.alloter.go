package mqc

import (
	"github.com/zhiyunliu/gel/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.processor.engine

	adapterEngine := server.NewAlloterEngine(engine,
		server.WithSrvType(e.Type()),
		server.WithSrvName(e.Name()),
		server.WithErrorEncoder(e.opts.encErr),
		server.WithRequestDecoder(e.opts.decReq),
		server.WithResponseEncoder(e.opts.encResp))

	server.RegistryEngineRoute(adapterEngine, e.opts.router)

}
