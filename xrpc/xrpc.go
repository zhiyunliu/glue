package xrpc

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"
)

//rpcResover 定义配置文件转换方法
type rpcResover interface {
	Name() string
	Resolve(setting config.Config) (Client, error)
}

var dbResolvers = make(map[string]rpcResover)

//Register 注册配置文件适配器
func Register(resolver rpcResover) {
	proto := resolver.Name()
	if _, ok := dbResolvers[proto]; ok {
		panic(fmt.Errorf("xrpc: 不能重复注册:%s", proto))
	}
	dbResolvers[proto] = resolver
}

//Deregister 清理配置适配器
func Deregister(name string) {
	delete(dbResolvers, name)
}

//newDB 根据适配器名称及参数返回配置处理器
func newXRPC(setting config.Config) (Client, error) {
	val := setting.Value("proto")
	proto := val.String()
	resolver, ok := dbResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("xrpc: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
