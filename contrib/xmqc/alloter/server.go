package alloter

import (
	"context"
	"fmt"

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

func newServer(srvcfg *serverConfig,
	router *engine.RouterGroup,
	opts ...engine.Option) (server *Server, err error) {

	server = &Server{
		srvCfg: srvcfg,
		engine: alloter.New(),
	}

	midwares, err := middleware.BuildMiddlewareList(srvcfg.Middlewares)
	if err != nil {
		err = fmt.Errorf("engine:[%s] BuildMiddlewareList,%w", srvcfg.Config.Proto, err)
		return
	}
	router.Use(midwares...)

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
	if e.processor != nil {
		return e.processor.Close()
	}
	return nil
}
