package caches

import (
	"fmt"

	"github.com/zhiyunliu/velocity/config"
)

type ICache interface {
	Name() string
	Get(key string) (string, error)
	Set(key string, val interface{}, expire int) error
	Del(key string) error
	HashGet(hk, key string) (string, error)
	HashDel(hk, key string) error
	Increase(key string) error
	Decrease(key string) error
	Expire(key string, expire int) error
	GetImpl() interface{}
}

//cacheResover 定义配置文件转换方法
type cacheResover interface {
	Name() string
	Resolve(setting *config.Setting) (ICache, error)
}

var cacheResolvers = make(map[string]cacheResover)

//RegisterCache 注册配置文件适配器
func RegisterCache(resolver cacheResover) {
	proto := resolver.Name()
	if _, ok := cacheResolvers[proto]; ok {
		panic(fmt.Errorf("cache: 不能重复注册:%s", proto))
	}
	cacheResolvers[proto] = resolver
}

//NewMQP 根据适配器名称及参数返回配置处理器
func NewCache(setting *config.Setting) (ICache, error) {
	proto := setting.GetProperty("proto")
	resolver, ok := cacheResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("cache: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
