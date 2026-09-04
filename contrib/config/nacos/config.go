package nacos

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zhiyunliu/glue/config"
)

type options struct {
	Group   string `json:"group,omitempty"`
	Cluster string `json:"cluster,omitempty"`
	DataID  string `json:"data_id,omitempty"`
}

type Config struct {
	opts   options
	client config_client.IConfigClient
}

func NewConfigSource(client config_client.IConfigClient, opts options) config.Source {
	return &Config{client: client, opts: opts}
}
func (c *Config) Name() string {
	return _name
}

func (f *Config) Path() string {
	return fmt.Sprintf("Group=%s&Cluster=%s&DataID=%s", f.opts.Group, f.opts.Cluster, f.opts.DataID)
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
