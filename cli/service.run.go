package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/zhiyunliu/gel/log"
)

func (p *ServiceApp) run() (err error) {
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
	p.register(ctx)
	return nil
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
		if err := p.options.Registrar.Register(rctx, instance); err != nil {
			return fmt.Errorf("register to %s [%s] error:%w", p.options.Registrar.Name(), p.options.Registrar.ServerConfigs(), err)
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
