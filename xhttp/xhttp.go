package xhttp

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

//httpResover 定义配置文件转换方法
type httpResover interface {
	Name() string
	Resolve(name string, setting config.Config) (Client, error)
}

var _resolvers = make(map[string]httpResover)

//Register 注册配置文件适配器
func Register(resolver httpResover) {
	proto := resolver.Name()
	if _, ok := _resolvers[proto]; ok {
		panic(fmt.Errorf("xhttp: 不能重复注册:%s", proto))
	}
	_resolvers[proto] = resolver
}

//Deregister 清理配置适配器
func Deregister(name string) {
	delete(_resolvers, name)
}

//newXhttp 根据适配器名称及参数返回配置处理器
func newXhttp(name string, setting config.Config) (Client, error) {
	val := setting.Value("proto")
	proto := val.String()
	if proto == "" {
		proto = "xhttp"
	}
	resolver, ok := _resolvers[proto]
	if !ok {
		return nil, fmt.Errorf("xhttp: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(name, setting)
}
