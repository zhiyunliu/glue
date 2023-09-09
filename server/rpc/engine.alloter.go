package rpc

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xrpc"
)

func (e *Server) resoverEngineRoute(server xrpc.Server) (err error) {
	adapterEngine, err := engine.NewEngine(e.opts.setting.Config.Engine, e.opts.config,
		engine.WithSrvType(e.Type()),
		engine.WithSrvName(e.Name()),
		engine.WithErrorEncoder(e.opts.encErr),
		engine.WithRequestDecoder(e.opts.decReq),
		engine.WithResponseEncoder(e.opts.encResp),
	)
	if err != nil {
		return
	}
	for _, m := range e.opts.setting.Middlewares {
		e.opts.router.Use(middleware.Resolve(&m))
	}

	engine.RegistryEngineRoute(adapterEngine, e.opts.router, e.opts.logOpts)
	return nil
}
