package mqc

import (
	"fmt"

	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

func (e *Server) resoverEngineRoute(processor *processor) (err error) {
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
	alloterEngine, ok := adapterEngine.GetImpl().(engine.AlloterEngine)
	if !ok {
		err = fmt.Errorf("engine:[%s] is not engine.AlloterEngine", e.opts.setting.Config.Engine)
		return
	}

	processor.engine = alloterEngine
	for _, m := range e.opts.setting.Middlewares {
		e.opts.router.Use(middleware.Resolve(&m))
	}

	engine.RegistryEngineRoute(adapterEngine, e.opts.router, e.opts.logOpts)
	return
}
