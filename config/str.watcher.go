package config

import (
	"context"
)

type strWatcher struct {
	f      *strSource
	ctx    context.Context
	cancel context.CancelFunc
}

var _ Watcher = (*strWatcher)(nil)

func newStrWatcher(f *strSource) (Watcher, error) {

	ctx, cancel := context.WithCancel(context.Background())
	return &strWatcher{f: f, ctx: ctx, cancel: cancel}, nil
}

func (w *strWatcher) Next() ([]*KeyValue, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	}
}

func (w *strWatcher) Stop() error {
	w.cancel()
	return nil
}
