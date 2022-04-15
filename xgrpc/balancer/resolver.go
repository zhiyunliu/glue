package balancer

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/zhiyunliu/gel/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

type registrarResolver struct {
	ctx            context.Context
	registrar      registry.Registrar
	reqPath        *url.URL
	clientConn     resolver.ClientConn
	waitGroup      sync.WaitGroup
	resolveNowChan chan struct{}
}

// ResolveNow resolves immediately
func (r *registrarResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case r.resolveNowChan <- struct{}{}:
	default:
	}
}

func (r *registrarResolver) Close() {
	r.waitGroup.Wait()
}

func (r *registrarResolver) watcher() {
	defer r.waitGroup.Done()

	if r.registrar == nil {
		return
	}
	go r.watchRegistrar()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.resolveNowChan:

		}
		instances, err := r.registrar.GetService(r.ctx, fmt.Sprintf("%s.%s", r.reqPath.Host, r.reqPath.Scheme))

		if err != nil {
			continue
		}

		addresses, err := r.buildAddress(instances)
		if err != nil || len(addresses) == 0 {
			continue
		}
		r.clientConn.UpdateState(resolver.State{Addresses: addresses})
	}
}

func (r *registrarResolver) buildAddress(instances []*registry.ServiceInstance) ([]resolver.Address, error) {

	var addresses = make([]resolver.Address, 0, len(instances))
	for _, v := range instances {
		for _, ep := range v.Endpoints {
			// ep=grpc://172.16.0.128:7080
			epv, _ := url.Parse(ep)

			a := resolver.Address{
				Addr:       epv.Host,
				ServerName: r.reqPath.Host,
				Attributes: attributes.New(v.Name, equalMap(v.Metadata)),
			}
			addresses = append(addresses, a)
		}
	}

	return addresses, nil
}

func (r *registrarResolver) watchRegistrar() {
	if r.registrar == nil {
		return
	}
	watcher, _ := r.registrar.Watch(r.ctx, fmt.Sprintf("%s.%s", r.reqPath.Host, r.reqPath.Scheme))
	for {

		select {
		case <-r.ctx.Done():
			return
		default:
			instances, err := watcher.Next()
			if err != nil {
				continue
			}
			addresses, err := r.buildAddress(instances)
			if err != nil || len(addresses) == 0 {
				continue
			}
			r.clientConn.UpdateState(resolver.State{Addresses: addresses})
		}
	}
}

type equalMap map[string]string

func (m equalMap) Equal(o interface{}) bool {
	mv, ok := o.(equalMap)
	if !ok {
		return false
	}
	if len(m) != len(mv) {
		return false
	}

	for k, v1 := range m {
		v2, ok := mv[k]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}
