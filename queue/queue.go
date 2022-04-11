package queue

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
	"github.com/zhiyunliu/golibs/xnet"
)

const queueTypeNode = "queues"

//StandardQueue queue
type StandardQueue struct {
	c container.Container
}

//NewStandardQueue 创建queue
func NewStandardQueue(c container.Container) *StandardQueue {
	return &StandardQueue{c: c}
}

//GetQueue GetQueue
func (s *StandardQueue) GetQueue(name string) (q IQueue) {
	obj, err := s.c.GetOrCreate(queueTypeNode, name, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(queueTypeNode).Value(name)
		cacheVal := cfgVal.String()
		//redis://localhost
		protoType, configName, err := xnet.Parse(cacheVal)
		if err != nil {
			panic(err)
		}
		queueCfg := cfg.Get(protoType).Get(configName)
		return newQueue(protoType, queueCfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(IQueue)
}
