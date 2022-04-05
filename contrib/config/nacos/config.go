package nacos

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/zhiyunliu/gel/config"
)

type options struct {
	Group  string `json:"group"`
	DataID string `json:"data_id"`
}

type Config struct {
	opts   options
	client config_client.IConfigClient
}

func NewConfigSource(client config_client.IConfigClient, opts options) config.Source {
	return &Config{client: client, opts: opts}
}

func (c *Config) Load() ([]*config.KeyValue, error) {
	content, err := c.client.GetConfig(vo.ConfigParam{
		DataId: c.opts.DataID,
		Group:  c.opts.Group,
	})
	if err != nil {
		return nil, err
	}
	k := c.opts.DataID
	return []*config.KeyValue{
		{
			Key:    k,
			Value:  []byte(content),
			Format: strings.TrimPrefix(filepath.Ext(k), "."),
		},
	}, nil
}

func (c *Config) Watch() (config.Watcher, error) {
	watcher := newWatcher(context.Background(), c.opts.DataID, c.opts.Group, c.client.CancelListenConfig)
	err := c.client.ListenConfig(vo.ConfigParam{
		DataId: c.opts.DataID,
		Group:  c.opts.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if dataId == watcher.dataID && group == watcher.group {
				watcher.content <- data
			}
		},
	})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

type Watcher struct {
	context.Context
	dataID             string
	group              string
	content            chan string
	cancelListenConfig cancelListenConfigFunc
	cancel             context.CancelFunc
}

type cancelListenConfigFunc func(params vo.ConfigParam) (err error)

func newWatcher(ctx context.Context, dataID string, group string, cancelListenConfig cancelListenConfigFunc) *Watcher {
	w := &Watcher{
		dataID:             dataID,
		group:              group,
		cancelListenConfig: cancelListenConfig,
		content:            make(chan string, 100),
	}
	ctx, cancel := context.WithCancel(ctx)
	w.Context = ctx
	w.cancel = cancel
	return w
}

func (w *Watcher) Next() ([]*config.KeyValue, error) {
	select {
	case <-w.Context.Done():
		return nil, nil
	case content := <-w.content:
		k := w.dataID
		return []*config.KeyValue{
			{
				Key:    k,
				Value:  []byte(content),
				Format: strings.TrimPrefix(filepath.Ext(k), "."),
			},
		}, nil
	}
}

func (w *Watcher) Close() error {
	err := w.cancelListenConfig(vo.ConfigParam{
		DataId: w.dataID,
		Group:  w.group,
	})
	w.cancel()
	return err
}

func (w *Watcher) Stop() error {
	return w.Close()
}
