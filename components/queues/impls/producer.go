package impls

import (
	"fmt"

	"github.com/zhiyunliu/velocity/config"
)

//IMQP 消息生产
type IMQP interface {
	Push(key string, value string) error
	Pop(key string) (string, error)
	Count(key string) (int64, error)
	Close() error
}

//imqpResover 定义配置文件转换方法
type imqpResover interface {
	Name() string
	Resolve(setting *config.Setting) (IMQP, error)
}

var mqpResolvers = make(map[string]imqpResover)

//RegisterProducer 注册配置文件适配器
func RegisterProducer(resolver imqpResover) {
	proto := resolver.Name()
	if _, ok := mqpResolvers[proto]; ok {
		panic(fmt.Errorf("mqp: 不能重复注册:%s", proto))
	}
	mqpResolvers[proto] = resolver
}

//NewMQP 根据适配器名称及参数返回配置处理器
func NewMQP(setting *config.Setting) (IMQP, error) {
	proto := setting.GetProperty("proto")
	resolver, ok := mqpResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("mqp: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
