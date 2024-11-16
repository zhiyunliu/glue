package queue

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

// MqpResover 定义配置文件转换方法
type MqpResolver interface {
	Name() string
	Resolve(setting config.Config, opts ...Option) (IMQP, error)
}

var mqpResolvers = make(map[string]MqpResolver)

// RegisterProducer 注册配置文件适配器
func RegisterProducer(resolver MqpResolver) {
	proto := resolver.Name()
	if _, ok := mqpResolvers[proto]; ok {
		panic(fmt.Errorf("mqp: 不能重复注册:%s", proto))
	}
	mqpResolvers[proto] = resolver
}

// Deregister 清理配置适配器
func DeregisterProducer(name string) {
	delete(mqpResolvers, name)
}

// NewMQP 根据适配器名称及参数返回配置处理器
func NewMQP(proto string, setting config.Config, opts ...Option) (IMQP, error) {
	resolver, ok := mqpResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("mqp: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting, opts...)
}
