package dlocker

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"
)

type DLocker interface {
	Acquire(expire int) (bool, error)
	Release() (bool, error)
	Renewal(expire int) error
}

type DLockerBuilder interface {
	Build(key string) DLocker
}

//cacheResover 定义配置文件转换方法
type xResover interface {
	Name() string
	Resolve(setting config.Config) (DLockerBuilder, error)
}

var lockerResolvers = make(map[string]xResover)

//RegisterCache 注册配置文件适配器
func Register(resolver xResover) {
	proto := resolver.Name()
	if _, ok := lockerResolvers[proto]; ok {
		panic(fmt.Errorf("dlocker: 不能重复注册:%s", proto))
	}
	lockerResolvers[proto] = resolver
}

//Deregister 清理配置适配器
func Deregister(name string) {
	delete(lockerResolvers, name)
}

//newCache 根据适配器名称及参数返回配置处理器
func newXLocker(proto string, setting config.Config) (DLockerBuilder, error) {
	resolver, ok := lockerResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("dlocker: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
