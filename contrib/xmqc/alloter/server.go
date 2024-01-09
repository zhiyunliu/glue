package alloter

import (
	"context"

	"github.com/zhiyunliu/glue/contrib/alloter"
	enginealloter "github.com/zhiyunliu/glue/contrib/engine/alloter"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/golibs/xnet"
)

const (
	Proto = "alloter"
)

type Server struct {
	srvCfg    *serverConfig
	engine    *alloter.Engine
	processor *processor
}

func newServer(cfg *serverConfig,
	router *engine.RouterGroup,
	opts ...engine.Option) (server *Server, err error) {

	server = &Server{
		srvCfg: cfg,
		engine: alloter.New(),
	}

	for _, m := range cfg.Middlewares {
		router.Use(middleware.Resolve(&m))
	}

	adapterEngine := enginealloter.NewAlloterEngine(server.engine, opts...)
	engine.RegistryEngineRoute(adapterEngine, router)
	return
}

func (e *Server) GetAddr() string {
	return e.srvCfg.Config.Addr
}

func (e *Server) GetProto() string {
	return Proto
}

func (e *Server) Serve(ctx context.Context) (err error) {
	//queue://default
	protoType, configName, err := xnet.Parse(e.GetAddr())
	if err != nil {
		return
	}
	//{"proto":"redis","addr":"redis://localhost"}
	cfg := global.Config.Get(protoType).Get(configName)

	protoType = cfg.Value("proto").String()
	e.processor, err = newProcessor(ctx, e.engine, protoType, configName, cfg)
	if err != nil {
		return
	}

	err = e.processor.Add(e.srvCfg.Tasks...)
	if err != nil {
		return
	}
	err = e.processor.Start()
	return err

}

func (e *Server) Stop(ctx context.Context) error {
	return e.processor.Close()
}
