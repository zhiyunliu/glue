package queue

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

// mqcResover 定义消息消费解析器
type MqcResover interface {
	Name() string
	Resolve(configName string, setting config.Config) (IMQC, error)
}

var mqcResolvers = make(map[string]MqcResover)

// RegisterConsumer 注册消息消费
func RegisterConsumer(resolver MqcResover) {
	name := resolver.Name()
	if _, ok := mqcResolvers[name]; ok {
		panic(fmt.Errorf("mqc: 不能重复注册:%s", name))
	}
	mqcResolvers[name] = resolver
}

// Deregister 清理配置适配器
func DeregisterrConsumer(name string) {
	delete(mqcResolvers, name)
}

// NewMQC 根据适配器名称及参数返回配置处理器
func NewMQC(proto, configName string, setting config.Config) (IMQC, error) {
	resolver, ok := mqcResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("mqc: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(configName, setting)
}
