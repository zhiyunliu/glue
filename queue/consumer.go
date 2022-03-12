package queue

import (
	"fmt"

	"github.com/zhiyunliu/velocity/config"
)

//mqcResover 定义消息消费解析器
type mqcResover interface {
	Name() string
	Resolve(setting config.Config) (IMQC, error)
}

var mqcResolvers = make(map[string]mqcResover)

//RegisterConsumer 注册消息消费
func RegisterConsumer(resolver mqcResover) {
	name := resolver.Name()
	if _, ok := mqcResolvers[name]; ok {
		panic(fmt.Errorf("mqc: 不能重复注册:%s", name))
	}
	mqcResolvers[name] = resolver
}

//NewMQC 根据适配器名称及参数返回配置处理器
func NewMQC(setting config.Config) (IMQC, error) {
	val := setting.Value("proto")
	proto, _ := val.String()
	resolver, ok := mqcResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("mqc: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
