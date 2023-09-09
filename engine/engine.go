package engine

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/context"
)

type AdapterEngine interface {
	NoMethod()
	NoRoute()
	Handle(method string, path string, callfunc HandlerFunc)
	Write(ctx context.Context, resp interface{})
	GetImpl() any
}
type HandlerFunc func(context.Context)

type Resover interface {
	Name() string
	Resolve(name string, config config.Config, opts ...Option) (AdapterEngine, error)
}

var engineResolvers = make(map[string]Resover)

// Register 注册配置文件适配器
func Register(resolver Resover) {
	proto := resolver.Name()
	if _, ok := engineResolvers[proto]; ok {
		panic(fmt.Errorf("engine: 不能重复注册:%s", proto))
	}
	engineResolvers[proto] = resolver
}

// Deregister 清理配置适配器
func Deregister(name string) {
	delete(engineResolvers, name)
}

// NewEngine 根据适配器名称及参数返回配置处理器
func NewEngine(proto string, setting config.Config, opts ...Option) (AdapterEngine, error) {
	resolver, ok := engineResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("engine: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(proto, setting, opts...)
}
