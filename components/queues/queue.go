package queues

import (
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/container"
)

const queueTypeNode = "queues"

//StandardQueue queue
type StandardQueue struct {
	c container.IContainer
}

//NewStandardQueue 创建queue
func NewStandardQueue(c container.IContainer) *StandardQueue {
	return &StandardQueue{c: c}
}

//GetQueue GetQueue
func (s *StandardQueue) GetQueue(name string) (q IQueue, err error) {
	obj, err := s.c.GetOrCreate(queueTypeNode, name, func(setting *config.Setting) (interface{}, error) {
		return newQueue(setting)
	})
	if err != nil {
		return nil, err
	}
	return obj.(IQueue), nil
}
