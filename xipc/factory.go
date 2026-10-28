package xipc

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
)

var clientResolvers = map[string]ClientResolver{}
var serverResolvers = map[string]ServerResolver{}

// Register 注册客户端协议实现。
func Register(resolver ClientResolver) {
	if resolver == nil {
		panic("ipc: nil client resolver")
	}
	name := resolver.Name()
	if _, ok := clientResolvers[name]; ok {
		panic(fmt.Errorf("ipc: duplicate client protocol: %s", name))
	}
	clientResolvers[name] = resolver
}

// Deregister 移除客户端协议实现。
func Deregister(name string) { delete(clientResolvers, name) }

// RegisterServer 注册服务端协议实现。
func RegisterServer(resolver ServerResolver) {
	if resolver == nil {
		panic("ipc: nil server resolver")
	}
	name := resolver.Name()
	if _, ok := serverResolvers[name]; ok {
		panic(fmt.Errorf("ipc: duplicate server protocol: %s", name))
	}
	serverResolvers[name] = resolver
}

// DeregisterServer 移除服务端协议实现。
func DeregisterServer(name string) { delete(serverResolvers, name) }

// NewClient 按协议创建客户端。
func NewClient(proto string, cfg config.Config) (Client, error) {
	resolver, ok := clientResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("ipc: unknown client protocol: %s", proto)
	}
	return resolver.Resolve(proto, cfg)
}

// NewServer 按协议创建服务端传输实现。
func NewServer(proto string, router *engine.RouterGroup, cfg config.Config, opts ...engine.Option) (Server, error) {
	resolver, ok := serverResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("ipc: unknown server protocol: %s", proto)
	}
	return resolver.Resolve(proto, router, cfg, opts...)
}
