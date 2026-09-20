package dlocker

import (
	"context"
	"errors"
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

// ErrLockLost 表示锁已过期、被删除或已由其他持有者获取。
var ErrLockLost = errors.New("dlocker: lock ownership lost")

type DLocker interface {
	//expire 秒
	Acquire(ctx context.Context, expire int) (bool, error)
	Release(ctx context.Context) (bool, error)
	//expire 秒
	Renewal(ctx context.Context, expire int) error
}

// FencedLocker 提供当前锁生命周期对应的 fencing token。
// 下游资源必须拒绝小于已处理 token 的写入，才能防止旧持有者继续写入。
type FencedLocker interface {
	DLocker
	FencingToken() uint64
}

// LossAwareLocker 提供自动续约期间的失锁通知，每次失锁最多写入一个错误。
type LossAwareLocker interface {
	DLocker
	Lost() <-chan error
}

type DLockerBuilder interface {
	Build(key string, opts ...Option) DLocker
}

// cacheResover 定义配置文件转换方法
type xResolver interface {
	Name() string
	Resolve(configName string, setting config.Config) (DLockerBuilder, error)
}

var lockerResolvers = make(map[string]xResolver)

// RegisterCache 注册配置文件适配器
func Register(resolver xResolver) {
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
