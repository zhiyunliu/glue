package appcli

import (
	"context"
	"time"

	"github.com/zhiyunliu/velocity/appcli/keys"
)

func (p *ServiceApp) run() (err error) {
	errChan := make(chan error)
	go func() {
		ctx := context.Background()
		ctx, p.CancelFunc = context.WithCancel(ctx)
		ctx = context.WithValue(ctx, keys.OptionsKey, p.options)
		err := p.manager.Start(ctx)
		errChan <- err
	}()

	select {
	case err = <-errChan:
		return err
	case <-time.After(time.Second):
		return nil
	}
}
