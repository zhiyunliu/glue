package balancer

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/zhiyunliu/glue/registry"
	"google.golang.org/grpc/resolver"
)

/*
// Builder creates a resolver that will be used to watch name resolution updates.
type Builder interface {
	// Build creates a new resolver for the given target.
	//
	// gRPC dial calls Build synchronously, and fails if the returned error is
	// not nil.
	Build(target Target, cc ClientConn, opts BuildOptions) (Resolver, error)
	// Scheme returns the scheme supported by this resolver.
	// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
	Scheme() string
}
*/

//ResolverBuilder creates a resolver that will be used to watch name resolution updates.
type registrarBuilder struct {
	ctx       context.Context
	registrar registry.Registrar
	reqPath   *url.URL
}

var _ resolver.Builder = &registrarBuilder{}

//NewRegistrarBuilder 新建builder
func NewRegistrarBuilder(ctx context.Context, registrar registry.Registrar, path *url.URL) resolver.Builder {

	builder := &registrarBuilder{
		ctx:       ctx,
		registrar: registrar,
		reqPath:   path,
	}

	return builder
}

// Scheme for mydns
func (b *registrarBuilder) Scheme() string {
	return b.reqPath.Scheme
}

// Build
func (b *registrarBuilder) Build(target resolver.Target, clientConn resolver.ClientConn, bopts resolver.BuildOptions) (resolver.Resolver, error) {

	rr := &registrarResolver{
		ctx:            b.ctx,
		registrar:      b.registrar,
		serviceName:    fmt.Sprintf("%s.%s", b.reqPath.Host, b.reqPath.Scheme),
		clientConn:     clientConn,
		waitGroup:      sync.WaitGroup{},
		resolveNowChan: make(chan struct{}, 1),
	}
	rr.waitGroup.Add(1)
	go rr.watcher()

	rr.ResolveNow(resolver.ResolveNowOptions{})
	return rr, nil
}
