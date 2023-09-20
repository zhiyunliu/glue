package dlocker

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

type DLocker interface {
	//expire 秒
	Acquire(expire int) (bool, error)
	Release() (bool, error)
	//expire 秒
	Renewal(expire int) error
}

type DLockerBuilder interface {
	Build(key string) DLocker
}

// cacheResover 定义配置文件转换方法
type xResover interface {
	Name() string
	Resolve(configName string, setting config.Config) (DLockerBuilder, error)
}

var lockerResolvers = make(map[string]xResover)

// RegisterCache 注册配置文件适配器
func Register(resolver xResover) {
	proto := resolver.Name()
	if _, ok := lockerResolvers[proto]; ok {
		panic(fmt.Errorf("dlocker: 不能重复注册:%s", proto))
	}
	lockerResolvers[proto] = resolver
}

// Deregister 清理配置适配器
func Deregister(name string) {
	delete(lockerResolvers, name)
}

// newCache 根据适配器名称及参数返回配置处理器
func newXLocker(proto, configName string, setting config.Config) (DLockerBuilder, error) {
	resolver, ok := lockerResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("dlocker: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(configName, setting)
}
