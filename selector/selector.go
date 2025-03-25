package selector

import (
	"context"

	"github.com/zhiyunliu/glue/errors"
)

// ErrNoAvailable is no available node.
var ErrNoAvailable = errors.InternalServer("no_available_node")

// Selector is node pick balancer.
type Selector interface {
	Rebalancer

	// Select nodes
	// if err == nil, selected and done must not be empty.
	Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error)
}

// Rebalancer is nodes rebalancer.
type Rebalancer interface {
	// Apply is apply all nodes when any changes happen
	Apply(nodes []Node)
}

// Builder build selector
type Builder interface {
	Name() string
	Build() Selector
}

type AddrInfo interface {
	Scheme() string
	Host() string
	Port() string
}

// Node is node interface.
type Node interface {
	// Address is the unique address under the same service
	Address() string

	// ServiceName is service name
	ServiceName() string

	// InitialWeight is the initial value of scheduling weight
	// if not set return nil
	InitialWeight() *int64

	// Version is service node version
	Version() string

	// Metadata is the kv pair metadata associated with the service instance.
	// version,namespace,region,protocol etc..
	Metadata() map[string]string
}

// DoneInfo is callback info when RPC invoke done.
type DoneInfo struct {
	// Response Error
	Err error
	// Response Metadata
	ReplyMeta ReplyMeta
}

// ReplyMeta is Reply Metadata.
type ReplyMeta interface {
	Get(key string) string
}

// DoneFunc is callback function when RPC invoke done.
type DoneFunc func(ctx context.Context, di DoneInfo)
