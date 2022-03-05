package queues

import (
	"fmt"

	"github.com/micro-plat/hydra/components/container"
	"github.com/micro-plat/hydra/conf"
	"github.com/micro-plat/lib4go/types"
)

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
	name := types.GetStringByIndex(names, 0, queueNameNode)
	obj, err := s.c.GetOrCreate(queueTypeNode, name, func(conf *conf.RawConf, keys ...string) (interface{}, error) {
		if conf.IsEmpty() {
			return nil, fmt.Errorf("节点/%s/%s未配置，或不可用", queueTypeNode, name)
		}
		return newQueue(conf.GetString("proto"), string(conf.GetRaw()))
	})
	if err != nil {
		return nil, err
	}
	return obj.(IQueue), nil
}
