package api

import (
	"fmt"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

func (e *Server) resoverEngineRoute() (err error) {
	adapterEngine, err := engine.NewEngine(e.opts.setting.Config.Engine, e.opts.config,
		engine.WithSrvType(e.Type()),
		engine.WithSrvName(e.Name()),
		engine.WithErrorEncoder(e.opts.encErr),
		engine.WithRequestDecoder(e.opts.decReq),
		engine.WithResponseEncoder(func(ctx context.Context, resp interface{}) error {
			for k, v := range e.opts.setting.Header {
				ctx.Response().Header(k, v)
			}
			return e.opts.encResp(ctx, resp)
		}))
	if err != nil {
		return
	}

	httpEngine, ok := adapterEngine.GetImpl().(engine.HttpEngine)
	if !ok {
		err = fmt.Errorf("engine:[%s] is not http.Handler", e.opts.setting.Config.Engine)
		return
	}
	e.opts.handler = httpEngine

	for _, m := range e.opts.setting.Middlewares {
		e.opts.router.Use(middleware.Resolve(&m))
	}

	for _, s := range e.opts.static {
		if s.IsFile {
			httpEngine.StaticFile(s.RouterPath, s.FilePath)
		} else {
			httpEngine.Static(s.RouterPath, s.FilePath)
		}
	}
	engine.RegistryEngineRoute(adapterEngine, e.opts.router, e.opts.logOpts)
	return nil
}
