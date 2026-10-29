package defaultipc

import (
	"context"

	"github.com/zhiyunliu/alloter"
	enginealloter "github.com/zhiyunliu/glue/contrib/engine/alloter"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xipc"

	"errors"
	"fmt"
	"sync"
)

type Server = xipc.Server

// Server 接收连接并分发已注册的命令。
type ipcServer struct {
	srvcfg *serverConfig

	closeErrs []error

	serveMu sync.Mutex
	serving bool
	closed  chan struct{}

	closeOnce sync.Once
	closeErr  error

	engine    *alloter.Engine
	processor *processor
}

// NewServer 创建并监听指定端点。
func newServer(srvcfg *serverConfig,
	router *engine.RouterGroup,
	opts ...engine.Option) (server *ipcServer, err error) {

	if srvcfg.Config.Addr == "" {
		srvcfg.Config.Addr = global.AppName
	}
	srvcfg.Config.Addr = xipc.GetIpcName(srvcfg.Config.Addr)

	server = &ipcServer{
		srvcfg: srvcfg,
		closed: make(chan struct{}),
		engine: alloter.New(),
	}
	server.processor, err = newProcessor(server.engine, srvcfg.Config.Capactiy, srvcfg.Config.Addr)
	if err != nil {
		return
	}

	midwares, err := middleware.BuildMiddlewareList(srvcfg.Middlewares)
	if err != nil {
		err = fmt.Errorf("engine:[%s] BuildMiddlewareList,%w", srvcfg.Config.Proto, err)
		return
	}
	router.Use(midwares...)

	adapterEngine := enginealloter.NewAlloterEngine(server.engine, opts...)
	engine.RegistryEngineRoute(adapterEngine, router)

	return server, nil
}

// Serve 阻塞接受连接，直到服务关闭或发生监听错误。
func (s *ipcServer) Serve(context.Context) error {
	s.serveMu.Lock()
	if s.serving {
		s.serveMu.Unlock()
		return ErrServerAlreadyServing
	}
	s.serving = true
	s.serveMu.Unlock()
	return s.processor.Start()
}

// Close 关闭监听器和活动连接，并等待连接处理协程退出。
func (s *ipcServer) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.processor.Stop()
	})
	return s.closeErr
}

// Address 返回服务端监听的原始端点。
func (s *ipcServer) Address() string {
	return s.processor.endpoint
}

func (s *ipcServer) GetProto() string           { return Proto }
func (s *ipcServer) GetAddr() string            { return s.Address() }
func (s *ipcServer) Stop(context.Context) error { return s.Close() }

var ErrServerAlreadyServing = errors.New("ipc: server is already serving")
