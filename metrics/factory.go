package metrics

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

type Provider interface {
	Name() string
	Counter() Counter
	Observer() Observer
	Gauge() Gauge
	GetImpl() interface{}
}

// resover 定义配置文件转换方法
type Resolver interface {
	Name() string
	Resolve(name string, config config.Config) (Provider, error)
}

var resolvers = make(map[string]Resolver)

// Register 注册配置文件适配器
func Register(resolver Resolver) {
	proto := resolver.Name()
	if _, ok := resolvers[proto]; ok {
		panic(fmt.Errorf("metrics: 不能重复注册:%s", proto))
	}
	resolvers[proto] = resolver
}

// Deregister 清理配置适配器
func Deregister(name string) {
	delete(resolvers, name)
}

// newProvider 根据适配器名称及参数返回配置处理器
func newProvider(proto string, setting config.Config) (Provider, error) {
	resolver, ok := resolvers[proto]
	if !ok {
		return nil, fmt.Errorf("metrics: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(proto, setting)
}
