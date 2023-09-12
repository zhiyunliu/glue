package cron

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

func (e *Server) resoverEngineRoute(processor *processor) (err error) {
	adapterEngine, err := engine.NewEngine(e.opts.srvCfg.Config.Engine, e.opts.config,
		engine.WithSrvType(e.Type()),
		engine.WithSrvName(e.Name()),
		engine.WithLogOptions(e.opts.logOpts),
		engine.WithErrorEncoder(e.opts.encErr),
		engine.WithRequestDecoder(e.opts.decReq),
		engine.WithResponseEncoder(e.opts.encResp),
	)
	if err != nil {
		return
	}

	processor.engine = adapterEngine

	for _, m := range e.opts.srvCfg.Middlewares {
		e.opts.router.Use(middleware.Resolve(&m))
	}

	engine.RegistryEngineRoute(adapterEngine, e.opts.router)
	return

}
