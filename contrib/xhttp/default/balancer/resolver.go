package balancer

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/zhiyunliu/gel/registry"
)

type Resovler interface {
	Pick() (string, error)
}

type registrarResolver struct {
	rwMutex        sync.RWMutex
	ctx            context.Context
	registrar      registry.Registrar
	reqPath        *url.URL
	addresses      []string //proto://ip:port
	waitGroup      sync.WaitGroup
	resolveNowChan chan struct{}
	next           int
}

func NewResolver(ctx context.Context, registrar registry.Registrar, reqPath *url.URL) Resovler {
	rr := &registrarResolver{
		ctx:            ctx,
		registrar:      registrar,
		reqPath:        reqPath,
		waitGroup:      sync.WaitGroup{},
		resolveNowChan: make(chan struct{}, 1),
	}
	rr.waitGroup.Add(1)
	go rr.watcher()
	rr.ResolveNow()
	return rr
}

// ResolveNow resolves immediately
func (r *registrarResolver) ResolveNow() {
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
		r.UpdateState(addresses)
	}
}

func (r *registrarResolver) UpdateState(addresses []string) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	r.addresses = addresses
}

func (r *registrarResolver) Pick() (string, error) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	if len(r.addresses) == 0 {
		return "", fmt.Errorf("no server address")
	}
	r.next = (r.next + 1) % len(r.addresses)
	return r.addresses[r.next], nil
}

func (r *registrarResolver) buildAddress(instances []*registry.ServiceInstance) ([]string, error) {

	var addresses = make([]string, 0, len(instances))
	for _, v := range instances {
		addresses = append(addresses, v.Endpoints...)
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
			r.UpdateState(addresses)
		}
	}
}
