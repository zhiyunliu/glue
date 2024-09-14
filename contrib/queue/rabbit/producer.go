package rabbit

import (
	"context"
	"sync"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/queue"
)

// Producer memcache配置文件
type Producer struct {
	client        *rabbitClient
	delayQueueMap *sync.Map
	closeChan     chan struct{}
	onceLock      sync.Once
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config, opts ...queue.Option) (m *Producer, err error) {
	m = &Producer{
		closeChan:     make(chan struct{}),
		delayQueueMap: &sync.Map{},
	}
	m.client, err = getRabbitClient(config, opts...)
	if err != nil {
		return
	}
	return
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) Push(ctx context.Context, key string, msg queue.Message) (err error) {
	return c.client.Publish(ctx, key, msg)
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) DelayPush(ctx context.Context, key string, msg queue.Message, delaySeconds int64) error {
	if delaySeconds <= 0 {
		return c.Push(ctx, key, msg)
	}

	return c.appendDelay(ctx, key, msg, delaySeconds)
}

// Close 释放资源
func (c *Producer) Close() error {
	c.onceLock.Do(func() {
		close(c.closeChan)
	})
	return c.client.Close()
}

type producerResolver struct {
}

func (s *producerResolver) Name() string {
	return Proto
}
func (s *producerResolver) Resolve(config config.Config, opts ...queue.Option) (queue.IMQP, error) {
	return NewProducer(config, opts...)
}
func init() {
	queue.RegisterProducer(&producerResolver{})
}
