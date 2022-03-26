package redis

import (
	"encoding/json"

	rds "github.com/go-redis/redis"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/contrib/redis"
	"github.com/zhiyunliu/velocity/queue"
)

// Producer memcache配置文件
type Producer struct {
	client *redis.Client
	config config.Config
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config) (m *Producer, err error) {
	m = &Producer{config: config}
	m.client, err = redis.NewByConfig(m.config)
	if err != nil {
		return
	}
	return
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) Push(key string, msg queue.Message) error {
	bytes, _ := json.Marshal(msg)
	_, err := c.client.RPush(key, bytesconv.BytesToString(bytes)).Result()
	return err
}

// Pop 移除并且返回 key 对应的 list 的第一个元素。
func (c *Producer) Pop(key string) (string, error) {
	r, err := c.client.LPop(key).Result()
	if err != nil && err == rds.Nil {
		return "", queue.Nil
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
func (s *producerResolver) Resolve(setting config.Config) (queue.IMQP, error) {
	return NewProducer(setting)
}
func init() {
	queue.RegisterProducer(&producerResolver{})
}
