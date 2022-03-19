package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

func (p *ServiceApp) run() (err error) {
	errChan := make(chan error)
	err = p.apprun(context.Background())
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
	instance, err := p.buildInstance()
	if err != nil {
		return err
	}
	eg, ctx := errgroup.WithContext(ctx)
	wg := sync.WaitGroup{}
	for _, srv := range p.options.Servers {
		srv := srv
		srv.Config(p.Config.Get(fmt.Sprintf("servers.%s", srv.Type())))
		eg.Go(func() error {
			<-ctx.Done() // wait for stop signal
			sctx, cancel := context.WithTimeout(NewContext(context.Background(), p), p.options.StopTimeout)
			defer cancel()
			return srv.Stop(sctx)
		})
		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			return srv.Start(ctx)
		})
	}
	wg.Wait()
	if p.options.Registrar != nil {
		rctx, rcancel := context.WithTimeout(ctx, p.options.RegistrarTimeout)
		defer rcancel()
		if err := p.options.Registrar.Register(rctx, instance); err != nil {
			return err
		}
		p.instance = instance
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
