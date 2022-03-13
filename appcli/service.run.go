package appcli

import (
	"time"
)

func (p *ServiceApp) run() (err error) {
	errChan := make(chan error)

	select {
	case err = <-errChan:
		return err
	case <-time.After(time.Second):
		return nil
	}
}

// func (p *ServiceApp) apprun(ctx context.Context) {

// 	eg, ctx := errgroup.WithContext(ctx)
// 	wg := sync.WaitGroup{}
// 	for _, srv := range p.opts.servers {
// 		srv := srv
// 		eg.Go(func() error {
// 			<-ctx.Done() // wait for stop signal
// 			sctx, cancel := context.WithTimeout(NewContext(context.Background(), p), p.opts.stopTimeout)
// 			defer cancel()
// 			return srv.Stop(sctx)
// 		})
// 		wg.Add(1)
// 		eg.Go(func() error {
// 			wg.Done()
// 			return srv.Start(ctx)
// 		})
// 	}
// 	wg.Wait()
// 	if p.opts.registrar != nil {
// 		rctx, rcancel := context.WithTimeout(p.opts.ctx, p.opts.registrarTimeout)
// 		defer rcancel()
// 		if err := p.opts.registrar.Register(rctx, instance); err != nil {
// 			return err
// 		}
// 		a.lk.Lock()
// 		a.instance = instance
// 		a.lk.Unlock()
// 	}
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, a.opts.sigs...)
// 	eg.Go(func() error {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return ctx.Err()
// 			case <-c:
// 				err := a.Stop()
// 				if err != nil {
// 					a.opts.logger.Errorf("failed to stop app: %v", err)
// 					return err
// 				}
// 			}
// 		}
// 	})
// 	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
// 		return err
// 	}
// }
