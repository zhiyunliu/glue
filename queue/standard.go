package queue

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
	"github.com/zhiyunliu/golibs/xnet"
)

const (
	TypeNode     = "queues"
	_defaultName = "default"
)

type StandardQueue interface {
	GetQueue(name ...string) (q IQueue)
}

//xQueue queue
type xQueue struct {
	c container.Container
}

//NewStandardQueue 创建queue
func NewStandardQueue(c container.Container) StandardQueue {
	return &xQueue{c: c}
}

//GetQueue GetQueue
func (s *xQueue) GetQueue(name ...string) (q IQueue) {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}

	obj, err := s.c.GetOrCreate(TypeNode, realName, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(TypeNode).Value(realName)
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

type xBuilder struct{}

func NewBuilder() container.StandardBuilder {
	return &xBuilder{}
}

func (xBuilder) Name() string {
	return TypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardQueue(c)
}
