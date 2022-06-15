package cache

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

type ICache interface {
	Name() string
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val interface{}, expire int) error
	Del(ctx context.Context, key string) error
	HashGet(ctx context.Context, hk, key string) (string, error)
	HashSet(ctx context.Context, hk, key string, val string) (bool, error)
	HashDel(ctx context.Context, hk, key string) error
	Increase(ctx context.Context, key string) (int64, error)
	Decrease(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expire int) error
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
