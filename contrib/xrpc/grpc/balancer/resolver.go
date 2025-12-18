package balancer

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

type registrarResolver struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	registrar      registry.Registrar
	serviceName    string //ip:port 或者服务名称
	clientConn     resolver.ClientConn
	waitGroup      *sync.WaitGroup
	resolveNowChan chan struct{}
	rwMutex        *sync.RWMutex
	lastSrvAddrs   []resolver.Address
}

func NewResolver(registrar registry.Registrar, serviceName string, clientConn resolver.ClientConn) resolver.Resolver {
	rr := &registrarResolver{
		registrar:      registrar,
		serviceName:    serviceName,
		clientConn:     clientConn,
		waitGroup:      &sync.WaitGroup{},
		rwMutex:        &sync.RWMutex{},
		resolveNowChan: make(chan struct{}, 1),
	}
	rr.ctx, rr.cancelFunc = context.WithCancel(context.Background())
	rr.doWatch()
	return rr
}

// ResolveNow resolves immediately
func (r *registrarResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	select {
	case r.resolveNowChan <- opts:
	default:
	}
}

func (r *registrarResolver) Close() {
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	r.waitGroup.Wait()
}

func (r *registrarResolver) doWatch() {
	address, ok := r.isServiceNameIpAddress()
	if ok {
		_ = r.clientConn.UpdateState(resolver.State{Addresses: address})
		return
	}
	if r.registrar == nil {
		return
	}

	go r.watchResolver()
	go r.watchRegistrar()
	go r.tickRefresh()
}

func (r *registrarResolver) watchResolver() {
	r.waitGroup.Add(1)
	defer func() {
		r.waitGroup.Done()
		log.Infof("grpc.watchResolver.exit.%s", r.serviceName)
	}()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.resolveNowChan:

		}
		instances, err := r.registrar.GetService(r.ctx, r.serviceName)
		if err != nil {
			log.Errorf("grpc:watchResolver.GetService=%s,error:%+v", r.serviceName, err)
			continue
		}

		addresses := r.buildAddress(instances)

		if !r.checkChange(addresses) {
			continue
		}

		err = r.clientConn.UpdateState(resolver.State{Addresses: addresses})
		if err != nil {
			log.Errorf("grpc:watchResolver.UpdateState=%s,error:%+v", r.serviceName, err)
		} else {
			r.updateLastSrvAddrs(addresses)
		}

	}
}

func (r *registrarResolver) watchRegistrar() {
	r.waitGroup.Add(1)
	defer func() {
		r.waitGroup.Done()
		log.Infof("grpc.watchRegistrar.exit.%s", r.serviceName)
	}()

	var (
		watcher registry.Watcher
		err     error
	)

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		watcher, err = r.registrar.Watch(r.ctx, r.serviceName)
		if err != nil {
			log.Errorf("grpc:watchRegistrar.Watch=%s.error:%+v", r.serviceName, err)
			time.Sleep(time.Second)
			continue
		}
		break
	}

	for {

		select {
		case <-r.ctx.Done():
			return
		default:
			instances, err := watcher.Next()
			if err != nil {
				log.Errorf("grpc:watchRegistrar.Next=%s,error:%+v", r.serviceName, err)
				time.Sleep(time.Second * 2)
				continue
			}
			addresses := r.buildAddress(instances)
			err = r.clientConn.UpdateState(resolver.State{Addresses: addresses})
			if err != nil {
				log.Errorf("grpc:watchRegistrar.UpdateState=%s,error:%+v", r.serviceName, err)
			} else {
				r.updateLastSrvAddrs(addresses)
			}
		}
	}
}

// 定时刷新
func (r registrarResolver) tickRefresh() {
	ticker := time.NewTicker(time.Second * 30) //30s刷新一次

	r.waitGroup.Add(1)
	defer func() {
		ticker.Stop()
		r.waitGroup.Done()
		log.Infof("grpc.tickRefresh.exit.%s", r.serviceName)
	}()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.resolveNowChan <- struct{}{}
		}
	}

}

// 检查服务是否存在变动
func (r *registrarResolver) checkChange(addresses []resolver.Address) bool {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()

	//服务数量不一致
	if len(addresses) != len(r.lastSrvAddrs) {
		return true
	}

	tmpMap := map[string]struct{}{}

	for _, addr := range r.lastSrvAddrs {
		tmpMap[addr.Addr] = struct{}{}
	}

	for _, addr := range addresses {
		if _, ok := tmpMap[addr.Addr]; !ok {
			return true
		}
	}
	//没有变动
	return false
}

func (r *registrarResolver) updateLastSrvAddrs(addresses []resolver.Address) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	r.lastSrvAddrs = addresses
}

func (r *registrarResolver) buildAddress(instances []*registry.ServiceInstance) []resolver.Address {
	var addresses = make([]resolver.Address, 0, len(instances))
	if len(instances) == 0 {
		return addresses
	}

	for _, v := range instances {
		if scheme, ok := v.Metadata["scheme"]; ok && !strings.EqualFold(scheme, "grpc") {
			continue
		}

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

	return addresses
}

func (r *registrarResolver) isServiceNameIpAddress() (address []resolver.Address, ok bool) {
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
