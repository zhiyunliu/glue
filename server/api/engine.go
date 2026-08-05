package api

import (
	"fmt"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

func (e *Server) resoverEngineRoute() (err error) {
	adapterEngine, err := engine.NewEngine(e.opts.srvCfg.Config.Engine, e.opts.config,
		engine.WithSrvType(e.Type()),
		engine.WithSrvName(e.Name()),
		engine.WithSvcName(e.ServiceName()),
		engine.WithLogOptions(e.opts.logOpts),
		engine.WithErrorEncoder(e.opts.encErr),
		engine.WithRequestDecoder(e.opts.decReq),
		engine.WithResponseEncoder(func(ctx context.Context, resp interface{}) error {
			for k, v := range e.opts.srvCfg.Header {
				ctx.Response().Header(k, v)
			}
			return e.opts.encResp(ctx, resp)
		}))
	if err != nil {
		return
	}

	httpEngine, ok := adapterEngine.GetImpl().(engine.HttpEngine)
	if !ok {
		err = fmt.Errorf("engine:[%s] is not http.Handler", e.opts.srvCfg.Config.Engine)
		return
	}
	e.opts.handler = httpEngine

	midwares, err := middleware.BuildMiddlewareList(e.opts.srvCfg.Middlewares)
	if err != nil {
		err = fmt.Errorf("engine:[%s] BuildMiddlewareList,%w", e.opts.srvCfg.Config.Engine, err)
		return
	}
	e.opts.router.Use(midwares...)

	for _, s := range e.opts.static {
		if s.FileSystem != nil {
			httpEngine.StaticFS(s.RouterPath, s.FileSystem)
			continue
		}
		if len(s.FilePath) > 0 {
			httpEngine.StaticFile(s.RouterPath, s.FilePath)
			continue
		}
		if len(s.DirPath) > 0 {
			httpEngine.Static(s.RouterPath, s.DirPath)
			continue
		}
	}
	engine.RegistryEngineRoute(adapterEngine, e.opts.router)
	return nil
}
