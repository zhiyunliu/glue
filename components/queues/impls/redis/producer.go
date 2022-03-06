package redis

import (
	rds "github.com/go-redis/redis"
	"github.com/micro-plat/hydra/components/queues/mq"
	"github.com/zhiyunliu/velocity/components/queues/impls"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/plugins/redis"
)

// Producer memcache配置文件
type Producer struct {
	servers []string
	client  *redis.Client
	setting *config.Setting
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(setting *config.Setting) (m *Producer, err error) {
	m = &Producer{setting: setting}
	m.client, err = redis.NewByConfig(m.setting)
	if err != nil {
		return
	}
	return
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) Push(key string, value string) error {
	_, err := c.client.RPush(key, value).Result()
	return err
}

// Pop 移除并且返回 key 对应的 list 的第一个元素。
func (c *Producer) Pop(key string) (string, error) {
	r, err := c.client.LPop(key).Result()
	if err != nil && err == rds.Nil {
		return "", mq.Nil
	}
	return r, err
}

// Count 获取列表中的元素个数
func (c *Producer) Count(key string) (int64, error) {
	return c.client.LLen(key).Result()
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
func (s *producerResolver) Resolve(setting *config.Setting) (impls.IMQP, error) {
	return NewProducer(setting)
}
func init() {
	impls.RegisterProducer(&producerResolver{})
}
