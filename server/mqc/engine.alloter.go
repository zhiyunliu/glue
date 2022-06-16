package mqc

import (
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.processor.engine

	adapterEngine := server.NewAlloterEngine(engine,
		server.WithSrvType(e.Type()),
		server.WithSrvName(e.Name()),
		server.WithErrorEncoder(e.opts.encErr),
		server.WithRequestDecoder(e.opts.decReq),
		server.WithResponseEncoder(e.opts.encResp))

	for _, m := range e.opts.setting.Middlewares {
		e.opts.router.Use(middleware.Resolve(&m))
	}

	server.RegistryEngineRoute(adapterEngine, e.opts.router)

}
