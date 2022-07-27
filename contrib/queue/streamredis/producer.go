package streamredis

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xtypes"
	redisqueue "github.com/zhiyunliu/redisqueue/v2"
)

// Producer memcache配置文件
type Producer struct {
	client   *redis.Client
	producer *redisqueue.Producer
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config) (m *Producer, err error) {
	m = &Producer{}
	m.client, err = getRedisClient(config)
	if err != nil {
		return
	}
	copts := &ProductOptions{}
	err = config.Scan(copts)
	if err != nil {
		return
	}

	opts := &redisqueue.ProducerOptions{
		RedisClient: m.client.UniversalClient,
	}
	if copts.StreamMaxLength > 0 {
		opts.StreamMaxLength = copts.StreamMaxLength
	}

	m.producer, err = redisqueue.NewProducerWithOptions(opts)
	return
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) Push(key string, msg queue.Message) error {
	vals := map[string]interface{}{
		"header": xtypes.SMap(msg.Header()),
		"body":   xtypes.XMap(msg.Body()),
	}
	//RPush(key, bytesconv.BytesToString(bytes)).Result()
	return c.producer.Enqueue(&redisqueue.Message{Stream: key, Values: vals})
}

// Count 获取列表中的元素个数
func (c *Producer) Count(key string) (int64, error) {
	return c.client.XLen(key).Result()
}

// Close 释放资源
func (c *Producer) Close() error {
	return c.client.Close()
}

type producerResolver struct {
}

func (s *producerResolver) Name() string {
	return Proto
}
func (s *producerResolver) Resolve(setting config.Config) (queue.IMQP, error) {
	return NewProducer(setting)
}
func init() {
	queue.RegisterProducer(&producerResolver{})
}
