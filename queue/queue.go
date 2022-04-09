package queue

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
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
	obj, err := s.c.GetOrCreate(queueTypeNode, name, func(setting config.Config) (interface{}, error) {

		return newQueue(setting)
	})
	if err != nil {
		panic(err)
	}
	return obj.(IQueue)
}
