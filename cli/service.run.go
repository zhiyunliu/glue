package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/opentelemetry"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/golibs/xnet"
)

func (p *ServiceApp) run() (err error) {

	errChan := make(chan error, 1)
	//p.svcCtx = context.Background()
	err = p.apprun()
	if err != nil {
		errChan <- err
	}
	select {
	case err = <-errChan:
		return err
	case <-time.After(time.Second):
		return nil
	}
}

func (p *ServiceApp) apprun() error {
	p.svcCtx = context.Background()
	if err := p.startingHooks(p.svcCtx); err != nil {
		return err
	}
	global.StartRunning()

	// 1. 初始化配置
	if err := opentelemetry.InitOtel(global.AppName, p.options.Config); err != nil {
		return err
	}

	p.closeWaitGroup.Add(len(p.options.Servers))
	for _, srv := range p.options.Servers {
		srv.Config(p.options.Config)
		err := srv.Start(context.Background())
		if err != nil {
			return err
		}
	}
	if err := p.startTraceServer(); err != nil {
		return err
	}

	if err := p.register(p.svcCtx); err != nil {
		return err
	}
	if err := p.startedHooks(p.svcCtx); err != nil {
		return err
	}

	return nil
}

func (p *ServiceApp) startTraceServer() error {
	if p.options.setting.TraceAddr == "" {
		log.Infof("pprof trace addr not set")
		return nil
	}

	errChan := make(chan error, 1)
	startTrace := func() {
		newAddr, err := xnet.GetAvaliableAddr(log.DefaultLogger, global.LocalIp, p.options.setting.TraceAddr)
		if err != nil {
			errChan <- err
			return
		}
		log.Infof("pprof trace config addr [%s]", p.options.setting.TraceAddr)
		log.Infof("pprof trace bind addr [%s]", newAddr)
		lsr, err := net.Listen("tcp", newAddr)
		if err != nil {
			errChan <- err
			return
		}
		traceSrv := &http.Server{}
		done := make(chan struct{})
		go func() {
			errChan <- traceSrv.Serve(lsr)
			close(done)
		}()

		select {
		case <-done:
			p.closeWaitGroup.Add(1)
		case <-time.After(time.Second):
			p.traceEndpoint, err = p.buildTraceEndpoint(newAddr)
			errChan <- err

		}
	}
	startTrace()

	err := <-errChan
	if err != nil {
		return fmt.Errorf("trace server error:%+v", err)
	}
	return nil
}

func (e *ServiceApp) buildTraceEndpoint(addr string) (item *registry.ServerItem, err error) {
	host, port, err := xnet.ExtractHostPort(addr)
	if err != nil {
		err = fmt.Errorf("trace Server Addr:%s 配置错误", addr)
		return
	}
	if host == "" {
		host = global.LocalIp
	}
	const scheme = "pprof"
	ep := transport.NewEndpoint(scheme, fmt.Sprintf("%s:%d", host, port))
	return &registry.ServerItem{
		ServiceName: fmt.Sprintf("%s-%s", scheme, global.AppName),
		EndpointURL: ep.String(),
	}, nil
}

func (p *ServiceApp) register(ctx context.Context) error {
	instance, err := p.buildInstance()
	if err != nil {
		return err
	}
	if p.options.Registrar != nil {
		log.Infof("register to %s [%s]", p.options.Registrar.Name(), p.options.Registrar.ServerConfigs())

		rctx, rcancel := context.WithTimeout(ctx, p.options.RegistrarTimeout)
		defer rcancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- p.options.Registrar.Register(rctx, instance)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return fmt.Errorf("register to %s [%s] error:%w", p.options.Registrar.Name(), p.options.Registrar.ServerConfigs(), err)
			}
			p.instance = instance
			log.Infof("register to %s completed", p.options.Registrar.Name())
		case <-rctx.Done():
			return fmt.Errorf("register to %s [%s] timeout:%s", p.options.Registrar.Name(), p.options.Registrar.ServerConfigs(), p.options.RegistrarTimeout.String())
		}
	}
	return nil
}

func (p *ServiceApp) deregister(ctx context.Context) error {
	if p.options.Registrar == nil {
		return nil
	}
	log.Infof("serviceApp close:%s unload registrar-%s", p.cliCtx.App.Name, p.options.Registrar.Name())
	if err := p.options.Registrar.Deregister(ctx, p.instance); err != nil {
		return err
	}
	return nil
}

func (p *ServiceApp) startingHooks(ctx context.Context) error {
	hooks := p.options.StartingHooks
	for i := range hooks {
		if err := hooks[i](ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *ServiceApp) startedHooks(ctx context.Context) error {
	hooks := p.options.StartedHooks
	for i := range hooks {
		if err := hooks[i](ctx); err != nil {
			return err
		}
	}
	return nil
}

type appKey struct{}

// NewContext returns a new Context that carries value.
func NewContext(ctx context.Context, s AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, s)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (s AppInfo, ok bool) {
	s, ok = ctx.Value(appKey{}).(AppInfo)
	return
}
