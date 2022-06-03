package balancer

import (
	"context"
	"sync"

	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/registry"
	"github.com/zhiyunliu/gel/selector"
)

type httpSelector struct {
	ctx         context.Context
	serviceName string
	registrar   registry.Registrar
	waitGroup   sync.WaitGroup
	selector    selector.Selector
}

func NewSelector(ctx context.Context, registrar registry.Registrar, serviceName, selectorName string) (selector.Selector, error) {
	tmpselector, err := selector.GetSelector(selectorName)
	if err != nil {
		return nil, err
	}

	rr := &httpSelector{
		ctx:         ctx,
		registrar:   registrar,
		serviceName: serviceName,
		waitGroup:   sync.WaitGroup{},
	}
	rr.selector = tmpselector
	rr.waitGroup.Add(1)
	go rr.watcher()
	rr.resolveNow()
	return rr, nil
}

func (r *httpSelector) Select(ctx context.Context, opts ...selector.SelectOption) (selected selector.Node, done selector.DoneFunc, err error) {
	return r.selector.Select(ctx, opts...)
}

func (r *httpSelector) Apply(nodes []selector.Node) {
	r.selector.Apply(nodes)
}

func (r *httpSelector) Close() {
	r.waitGroup.Wait()
}

// ResolveNow resolves immediately
func (r *httpSelector) resolveNow() {
	instances, err := r.registrar.GetService(r.ctx, r.serviceName)
	if err != nil {
		log.Errorf("http:registrar.GetService=%s,error:%+v", r.serviceName, err)
		return
	}
	r.Apply(r.buildAddress(instances))
}

func (r *httpSelector) buildAddress(instances []*registry.ServiceInstance) []selector.Node {

	var addresses = make([]selector.Node, 0, len(instances))
	for _, v := range instances {
		for _, ep := range v.Endpoints {
			addresses = append(addresses, selector.NewNode(ep, v))
		}
	}
	return addresses
}

func (r *httpSelector) watchRegistrar() {
	if r.registrar == nil {
		return
	}
	watcher, _ := r.registrar.Watch(r.ctx, r.serviceName)
	for {

		select {
		case <-r.ctx.Done():
			return
		default:
			instances, err := watcher.Next()
			if err != nil {
				log.Errorf("http:registrar.Next=%s,error:%+v", r.serviceName, err)
				continue
			}
			addresses := r.buildAddress(instances)
			r.Apply(addresses)
		}
	}
}

func (r *httpSelector) watcher() {
	defer r.waitGroup.Done()

	if r.registrar == nil {
		return
	}
	go r.watchRegistrar()

	<-r.ctx.Done()
}
