package nacos

import (
	"context"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zhiyunliu/glue/registry"
)

var _ registry.Watcher = (*watcher)(nil)

type watcher struct {
	serviceName string
	clusters    []string
	groupName   string
	ctx         context.Context
	cancel      context.CancelFunc
	watchChan   chan struct{}
	cli         naming_client.INamingClient
}

func newWatcher(ctx context.Context, cli naming_client.INamingClient, serviceName, groupName string, clusters []string) (*watcher, error) {
	w := &watcher{
		serviceName: serviceName,
		clusters:    clusters,
		groupName:   groupName,
		cli:         cli,
		watchChan:   make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)

	e := w.cli.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		Clusters:    clusters,
		GroupName:   groupName,
		SubscribeCallback: func(services []model.Instance, err error) {
			select {
			case w.watchChan <- struct{}{}:
			default:
			}
		},
	})
	return w, e
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case <-w.watchChan:
	}
	res, err := w.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: w.serviceName,
		GroupName:   w.groupName,
		Clusters:    w.clusters,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}
	items := buildServiceInstanceList(w.serviceName, res)
	return items, nil
}

func (w *watcher) Stop() error {
	w.cancel()
	return w.cli.Unsubscribe(&vo.SubscribeParam{
		ServiceName: w.serviceName,
		GroupName:   w.groupName,
		Clusters:    w.clusters,
	})
}
