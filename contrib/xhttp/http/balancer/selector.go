package balancer

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/selector"
)

type httpSelector struct {
	ctx         context.Context
	serviceName string
	registrar   registry.Registrar
	selector    selector.Selector

	waitGroup      *sync.WaitGroup
	resolveNowChan chan struct{}
}

var _ selector.Selector = (*httpSelector)(nil)

func NewSelector(ctx context.Context, registrar registry.Registrar, reqPath *url.URL, selectorName string) (selector.Selector, error) {
	tmpselector, err := selector.GetSelector(selectorName)
	if err != nil {
		return nil, err
	}

	rr := &httpSelector{
		ctx:            ctx,
		registrar:      registrar,
		serviceName:    reqPath.Host,
		resolveNowChan: make(chan struct{}, 1),
		waitGroup:      &sync.WaitGroup{},
	}
	rr.selector = tmpselector
	if strings.EqualFold(reqPath.Scheme, "xhttp") {
		rr.doWatch()
		rr.resolveNow()
	} else {
		rr.Apply(rr.buildOriginAddress(reqPath))
	}
	return rr, nil
}

func (r *httpSelector) ServiceName() string {
	return r.serviceName
}

func (r *httpSelector) Select(ctx context.Context, opts ...selector.SelectOption) (selected selector.Node, done selector.DoneFunc, err error) {
	return r.selector.Select(ctx, opts...)
}

func (r *httpSelector) Apply(nodes []selector.Node) {
	r.selector.Apply(nodes)
}

func (r *httpSelector) Nodes() (nodes []selector.Node) {
	return r.selector.Nodes()
}

// resolveNow resolves immediately
func (r *httpSelector) resolveNow() {
	r.resolveNowChan <- struct{}{}
}

func (r *httpSelector) buildOriginAddress(reqPath *url.URL) []selector.Node {
	var addresses = make([]selector.Node, 0, 1)
	addresses = append(addresses, &node{
		addr:        fmt.Sprintf("%s://%s", reqPath.Scheme, reqPath.Host),
		serviceName: reqPath.Scheme,
	})
	return addresses
}

func (r *httpSelector) buildAddress(instances []*registry.ServiceInstance) []selector.Node {

	var addresses = make([]selector.Node, 0, len(instances))
	for _, v := range instances {
		if scheme, ok := v.Metadata["scheme"]; ok && !strings.EqualFold(scheme, "http") {
			continue
		}
		for _, ep := range v.Endpoints {
			addresses = append(addresses, selector.NewNode(ep, v))
		}
	}
	return addresses
}

func (r *httpSelector) doWatch() {
	if r.registrar == nil {
		return
	}
	go r.watchResolver()
	go r.watchRegistrar()
	go r.tickRefresh()
}

func (r *httpSelector) watchResolver() {
	r.waitGroup.Add(1)
	defer func() {
		r.waitGroup.Done()
		log.Infof("xhttp.watchResolver.exit.%s", r.serviceName)
	}()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.resolveNowChan:

		}
		instances, err := r.registrar.GetService(r.ctx, r.serviceName)
		if err != nil {
			log.Errorf("xhttp:watchResolver.GetService=%s,error:%+v", r.serviceName, err)
			continue
		}
		addresses := r.buildAddress(instances)

		if !r.checkChange(addresses) {
			continue
		}
		r.Apply(addresses)
	}
}

func (r *httpSelector) watchRegistrar() {
	if r.registrar == nil {
		return
	}

	var (
		watcher registry.Watcher
		err     error
	)
	for {
		watcher, err = r.registrar.Watch(r.ctx, r.serviceName)
		if err != nil {
			log.Errorf("xhttp:watchRegistrar.Watch=%s.error:%+v", r.serviceName, err)
			time.Sleep(time.Second * 2)
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
				log.Errorf("xhttp:watchRegistrar.Next=%s,error:%+v", r.serviceName, err)
				time.Sleep(time.Second * 2)
				continue
			}
			addresses := r.buildAddress(instances)
			r.Apply(addresses)
		}
	}
}

func (r *httpSelector) tickRefresh() {
	ticker := time.NewTicker(time.Second * 30) //30s刷新一次

	r.waitGroup.Add(1)
	defer func() {
		ticker.Stop()
		r.waitGroup.Done()
		log.Infof("xhttp.tickRefresh.exit.%s", r.serviceName)
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
func (r *httpSelector) checkChange(addresses []selector.Node) bool {
	//服务数量不一致
	if len(addresses) != len(r.Nodes()) {
		return true
	}

	tmpMap := map[string]struct{}{}

	for _, addr := range r.Nodes() {
		tmpMap[addr.Address()] = struct{}{}
	}

	for _, addr := range addresses {
		if _, ok := tmpMap[addr.Address()]; !ok {
			return true
		}
	}
	//没有变动
	return false
}
