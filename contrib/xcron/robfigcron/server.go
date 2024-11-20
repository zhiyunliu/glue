package robfigcron

import (
	"context"

	"github.com/zhiyunliu/alloter"
	enginealloter "github.com/zhiyunliu/glue/contrib/engine/alloter"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xcron"
)

const (
	Proto = "robfigcron"
)

type Server struct {
	srvCfg    *serverConfig
	engine    *alloter.Engine
	processor *processor
	router    *engine.RouterGroup
}

func newServer(srvcfg *serverConfig,
	router *engine.RouterGroup,
	opts ...engine.Option) (server *Server, err error) {

	server = &Server{
		srvCfg: srvcfg,
		router: router,
		engine: alloter.New(),
	}

	var midwares []middleware.Middleware
	for _, m := range srvcfg.Middlewares {
		midware, ierr := middleware.Resolve(&m)
		if ierr != nil {
			err = ierr
			return
		}
		midwares = append(midwares, midware)

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
	e.processor, err = newProcessor(ctx, e.engine)
	if err != nil {
		return
	}
	err = e.processor.Add(e.srvCfg.Jobs...)
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

func (e *Server) AddJob(jobs ...*xcron.Job) (keys []string, err error) {
	keys = make([]string, len(jobs))
	for i := range jobs {
		keys[i] = jobs[i].GetKey()
	}
	err = e.processor.Add(jobs...)
	return
}

func (e *Server) RemoveJob(keys ...string) {
	for i := range keys {
		e.processor.Remove(keys[i])
	}
}
