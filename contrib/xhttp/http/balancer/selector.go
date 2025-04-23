package balancer

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/selector"
)

type SelectorWrapper interface {
	selector.Selector
	ServiceName() string
}
type httpSelector struct {
	ctx         context.Context
	serviceName string
	registrar   registry.Registrar
	selector    selector.Selector
}

var _ SelectorWrapper = (*httpSelector)(nil)

func NewSelector(ctx context.Context, registrar registry.Registrar, reqPath *url.URL, selectorName string) (SelectorWrapper, error) {
	tmpselector, err := selector.GetSelector(selectorName)
	if err != nil {
		return nil, err
	}

	rr := &httpSelector{
		ctx:         ctx,
		registrar:   registrar,
		serviceName: reqPath.Host,
	}
	rr.selector = tmpselector
	if strings.EqualFold(reqPath.Scheme, "xhttp") {
		go rr.watcher()
		rr.resolveNow()
	} else {
		rr.Apply(rr.buildAddress(reqPath))
	}

	return rr, nil
}

func (r *httpSelector) Select(ctx context.Context, opts ...selector.SelectOption) (selected selector.Node, done selector.DoneFunc, err error) {
	return r.selector.Select(ctx, opts...)
}

func (r *httpSelector) Apply(nodes []selector.Node) {
	r.selector.Apply(nodes)
}

func (r *httpSelector) ServiceName() string {
	return r.serviceName
}

// resolveNow resolves immediately
func (r *httpSelector) resolveNow() {
	if r.registrar == nil {
		return
	}
	instances, err := r.registrar.GetService(r.ctx, r.serviceName)
	if err != nil {
		log.Errorf("http:registrar.GetService=%s,error:%s", r.serviceName, err.Error())
		return
	}
	r.Apply(r.buildServiceNode(instances))
}

func (r *httpSelector) buildAddress(reqPath *url.URL) []selector.Node {
	var addresses = make([]selector.Node, 0, 1)
	addresses = append(addresses, &node{
		addr:        fmt.Sprintf("%s://%s", reqPath.Scheme, reqPath.Host),
		serviceName: reqPath.Scheme,
	})
	return addresses
}

func (r *httpSelector) buildServiceNode(instances []*registry.ServiceInstance) []selector.Node {

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

func (r *httpSelector) watcher() {
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
			addresses := r.buildServiceNode(instances)
			r.Apply(addresses)
		}
	}
}
