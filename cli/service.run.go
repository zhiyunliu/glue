package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/golibs/xlog"
)

func (p *ServiceApp) run() (err error) {

	if p.cliCtx.Bool("nostd") {
		xlog.RemoveAppender(xlog.Stdout)
	}

	errChan := make(chan error)
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
	errChan := make(chan error)
	go func() {
		lsr, err := net.Listen("tcp", p.options.setting.TraceAddr)
		if err != nil {
			errChan <- err
			log.Errorf("start trace server listen error:%+v", err)
			return
		}
		traceSrv := &http.Server{}

		if err = traceSrv.Serve(lsr); err != nil {
			errChan <- err
			log.Errorf("start trace server Serve error:%+v", err)
			return
		}
		errChan <- nil
		<-p.svcCtx.Done()
	}()

	err := <-errChan
	return err
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

		errChan := make(chan error)
		go func() {
			if err := p.options.Registrar.Register(rctx, instance); err != nil {
				errChan <- err
			}
		}()

		select {
		case err := <-errChan:
			return fmt.Errorf("register to %s [%s] error:%w", p.options.Registrar.Name(), p.options.Registrar.ServerConfigs(), err)
		case <-rctx.Done():
			return fmt.Errorf("register to %s [%s] timeout:%d", p.options.Registrar.Name(), p.options.Registrar.ServerConfigs(), p.options.RegistrarTimeout)
		}

		p.instance = instance
		log.Infof("register to %s completed", p.options.Registrar.Name())
	}
	return nil
}

func (p *ServiceApp) deregister(ctx context.Context) error {

	if p.options.Registrar == nil {
		return nil
	}
	log.Infof("deregister %s", p.options.Registrar.Name())
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
