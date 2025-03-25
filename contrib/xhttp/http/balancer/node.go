package balancer

import (
	"net/url"

	"github.com/zhiyunliu/glue/selector"
)

var (
	_ selector.Node     = (*node)(nil)
	_ selector.AddrInfo = (*node)(nil)
)

type node struct {
	addr        string
	serviceName string
	addrUrl     *url.URL
}

func (n node) Address() string {
	return n.addr
}

// ServiceName is service name
func (n node) ServiceName() string {
	return n.serviceName
}

// InitialWeight is the initial value of scheduling weight
// if not set return nil
func (n node) InitialWeight() *int64 {
	return nil
}

// Version is service node version
func (n node) Version() string {
	return "0.0.0.0"
}

// Metadata is the kv pair metadata associated with the service instance.
// version,namespace,region,protocol etc..
func (n node) Metadata() map[string]string {
	return map[string]string{}
}

func (n *node) Scheme() string {
	n.init()
	return n.addrUrl.Scheme
}

func (n *node) Host() string {
	n.init()
	return n.addrUrl.Hostname()
}

func (n *node) Port() string {
	n.init()
	return n.addrUrl.Port()
}

func (n *node) init() {
	if n.addrUrl == nil {
		n.addrUrl, _ = url.Parse(n.addr)
	}
}
