package streamredis

import (
	"sync"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xtypes"
	redisqueue "github.com/zhiyunliu/redisqueue/v2"
)

const (
	DELAY_QUEUE_NAME = "glue:delayqueue:stream"
)

// Producer memcache配置文件
type Producer struct {
	client        *redis.Client
	producer      *redisqueue.Producer
	opts          *ProductOptions
	closeChan     chan struct{}
	delayQueueMap *sync.Map
	onceLock      sync.Once
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config, opts ...queue.Option) (m *Producer, err error) {

	m = &Producer{
		closeChan:     make(chan struct{}),
		delayQueueMap: &sync.Map{},
	}
	m.client, err = getRedisClient(config, opts...)
	if err != nil {
		return
	}
	copts := &ProductOptions{
		DelayInterval: 2,
	}
	err = config.Scan(copts)
	if err != nil {
		return
	}
	m.opts = copts

	pdtOpts := &redisqueue.ProducerOptions{
		StreamMaxLength:      10000,
		RedisClient:          m.client.UniversalClient,
		ApproximateMaxLength: copts.ApproximateMaxLength,
	}
	if copts.StreamMaxLength > 0 {
		pdtOpts.StreamMaxLength = copts.StreamMaxLength
	}

	m.producer, err = redisqueue.NewProducerWithOptions(pdtOpts)
	return
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) Push(key string, msg queue.Message) error {
	vals := map[string]interface{}{
		"header": xtypes.SMap(msg.Header()),
		"body":   msg.Body(),
	}
	//RPush(key, bytesconv.BytesToString(bytes)).Result()
	return c.producer.Enqueue(&redisqueue.Message{Stream: key, Values: vals})
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) DelayPush(key string, msg queue.Message, delaySeconds int64) error {
	if delaySeconds <= 0 {
		return c.Push(key, msg)
	}

	return c.appendDelay(key, msg, delaySeconds)
}

// Count 获取列表中的元素个数
func (c *Producer) Count(key string) (int64, error) {
	return c.client.XLen(key).Result()
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
func (s *producerResolver) Resolve(setting config.Config, opts ...queue.Option) (queue.IMQP, error) {
	return NewProducer(setting, opts...)
}
func init() {
	queue.RegisterProducer(&producerResolver{})
}
