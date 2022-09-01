package queue

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
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
	if realName == "" {
		realName = _defaultName
	}

	obj, err := s.c.GetOrCreate(TypeNode, realName, func(cfg config.Config) (interface{}, error) {
		//{"proto":"redis","addr":"redis://localhost"}
		cfgVal := cfg.Get(TypeNode).Get(realName)
		protoType := cfgVal.Value("proto").String()
		return newQueue(protoType, cfgVal)
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
