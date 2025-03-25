package selector

import (
	"net/url"
	"strconv"

	"github.com/zhiyunliu/glue/registry"
)

var _ Node = (*DefaultNode)(nil)
var _ AddrInfo = (*DefaultNode)(nil)

// DefaultNode is selector node
type DefaultNode struct {
	addr     string
	weight   *int64
	version  string
	name     string
	metadata map[string]string
	addrUrl  *url.URL
}

// Address is node address
func (n *DefaultNode) Address() string {
	return n.addr
}

// ServiceName is node serviceName
func (n *DefaultNode) ServiceName() string {
	return n.name
}

// InitialWeight is node initialWeight
func (n *DefaultNode) InitialWeight() *int64 {
	return n.weight
}

// Version is node version
func (n *DefaultNode) Version() string {
	return n.version
}

// Metadata is node metadata
func (n *DefaultNode) Metadata() map[string]string {
	return n.metadata
}

func (n *DefaultNode) Scheme() string {
	n.init()
	return n.addrUrl.Scheme
}

func (n *DefaultNode) Host() string {
	n.init()
	return n.addrUrl.Hostname()
}

func (n *DefaultNode) Port() string {
	n.init()
	return n.addrUrl.Port()
}

func (n *DefaultNode) init() {
	if n.addrUrl == nil {
		n.addrUrl, _ = url.Parse(n.addr)
	}
}

// NewNode new node
func NewNode(item registry.ServerItem, ins *registry.ServiceInstance) Node {
	n := &DefaultNode{
		addr: item.EndpointURL,
	}
	if ins != nil {
		n.name = ins.Name
		n.version = ins.Version
		n.metadata = ins.Metadata
		if str, ok := ins.Metadata["weight"]; ok {
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				n.weight = &weight
			}
		}
	}
	return n
}
