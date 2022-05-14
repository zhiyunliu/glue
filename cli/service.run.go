package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/golibs/xlog"
)

func (p *ServiceApp) run() (err error) {

	if p.cliCtx.Bool("nostd") {
		xlog.RemoveAppender(xlog.Stdout)
	}

	errChan := make(chan error, 1)
	p.svcCtx = context.Background()
	err = p.apprun(p.svcCtx)
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

func (p *ServiceApp) apprun(ctx context.Context) error {
	p.closeWaitGroup.Add(len(p.options.Servers))
	for _, srv := range p.options.Servers {
		srv.Config(p.options.Config)
		err := srv.Start(ctx)
		if err != nil {
			return err
		}
	}
	if err := p.register(ctx); err != nil {
		return err
	}
	if err := p.startTraceServer(); err != nil {
		return err
	}
	return nil
}

func (p *ServiceApp) startTraceServer() error {
	errChan := make(chan error, 1)
	go func() {
		log.Infof("pprof trace addr %s%s", global.LocalIp, p.options.setting.TraceAddr)
		lsr, err := net.Listen("tcp", p.options.setting.TraceAddr)
		if err != nil {
			errChan <- err
			log.Errorf("trace server listen error:%+v", err)
			return
		}
		traceSrv := &http.Server{}
		select {
		case errChan <- traceSrv.Serve(lsr):
			p.closeWaitGroup.Add(1)
			return
		case <-p.svcCtx.Done():
			return
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("trace server Serve error:%+v", err)
	case <-time.After(time.Second):
		return nil
	}
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
