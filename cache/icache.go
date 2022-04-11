package cache

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"
)

type ICache interface {
	Name() string
	Get(key string) (string, error)
	Set(key string, val interface{}, expire int) error
	Del(key string) error
	HashGet(hk, key string) (string, error)
	HashSet(hk, key string, val string) (bool, error)
	HashDel(hk, key string) error
	Increase(key string) (int64, error)
	Decrease(key string) (int64, error)
	Expire(key string, expire int) error
	GetImpl() interface{}
}

//cacheResover 定义配置文件转换方法
type cacheResover interface {
	Name() string
	Resolve(setting config.Config) (ICache, error)
}

var cacheResolvers = make(map[string]cacheResover)

//RegisterCache 注册配置文件适配器
func Register(resolver cacheResover) {
	proto := resolver.Name()
	if _, ok := cacheResolvers[proto]; ok {
		panic(fmt.Errorf("cache: 不能重复注册:%s", proto))
	}
	cacheResolvers[proto] = resolver
}

//Deregister 清理配置适配器
func Deregister(name string) {
	delete(cacheResolvers, name)
}

//newCache 根据适配器名称及参数返回配置处理器
func newCache(proto string, setting config.Config) (ICache, error) {
	resolver, ok := cacheResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("cache: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
