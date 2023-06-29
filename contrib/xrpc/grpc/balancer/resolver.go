package balancer

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

type registrarResolver struct {
	ctx            context.Context
	registrar      registry.Registrar
	serviceName    string //ip:port 或者服务名称
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

	address, ok := r.isServiceNameIpAddress()
	if ok {
		r.clientConn.UpdateState(resolver.State{Addresses: address})
		return
	}

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
		instances, err := r.registrar.GetService(r.ctx, r.serviceName)
		if err != nil {
			log.Errorf("grpc:registrar.GetService=%s,error:%+v", r.serviceName, err)

			continue
		}

		address, err = r.buildAddress(instances)
		if err != nil || len(address) == 0 {
			continue
		}
		r.clientConn.UpdateState(resolver.State{Addresses: address})
	}
}

func (r *registrarResolver) buildAddress(instances []*registry.ServiceInstance) ([]resolver.Address, error) {

	var addresses = make([]resolver.Address, 0, len(instances))
	for _, v := range instances {
		for _, ep := range v.Endpoints {
			// ep=grpc://172.16.0.128:7080
			epv, _ := url.Parse(ep.EndpointURL)

			a := resolver.Address{
				Addr:       epv.Host,
				ServerName: v.Name,
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
	watcher, _ := r.registrar.Watch(r.ctx, r.serviceName)
	for {

		select {
		case <-r.ctx.Done():
			return
		default:
			instances, err := watcher.Next()
			if err != nil {
				log.Errorf("grpc:registrar.Watch=%s,error:%+v", r.serviceName, err)

				continue
			}
			addresses, err := r.buildAddress(instances)
			if err != nil {

				continue
			}
			r.clientConn.UpdateState(resolver.State{Addresses: addresses})
		}
	}
}

func (r *registrarResolver) isServiceNameIpAddress() (address []resolver.Address, ok bool) {
	ok = false
	parties := strings.SplitN(r.serviceName, ":", 2)
	if len(parties) <= 1 {
		ok = false
		return
	}

	address = make([]resolver.Address, 0, 1)
	a := resolver.Address{
		Addr:       r.serviceName,
		ServerName: r.serviceName,
		Attributes: attributes.New(r.serviceName, equalMap(map[string]string{})),
	}
	address = append(address, a)
	ok = true
	return
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
