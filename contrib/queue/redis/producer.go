package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	rds "github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
)

// Producer memcache配置文件
type Producer struct {
	opts      *ProductOptions
	client    *redis.Client
	closeChan chan struct{}
	onceLock  sync.Once
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config) (m *Producer, err error) {
	m = &Producer{}
	m.client, err = getRedisClient(config)
	if err != nil {
		return
	}
	m.opts = &ProductOptions{
		DelayQueueName: "glue:delayqueue:list",
		RangeSeconds:   1800,
		DelayInterval:  5,
	}
	err = config.Scan(m.opts)
	if err != nil {
		return
	}
	return
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) Push(key string, msg queue.Message) error {
	bytes, _ := json.Marshal(map[string]interface{}{
		"header": msg.Header(),
		"body":   msg.Body(),
	})

	_, err := c.client.RPush(key, bytesconv.BytesToString(bytes)).Result()
	return err
}

// Push 向存于 key 的列表的尾部插入所有指定的值
func (c *Producer) DelayPush(key string, msg queue.Message, delaySeconds int64) error {
	if delaySeconds <= 0 {
		return c.Push(key, msg)
	}

	bytes, _ := json.Marshal(map[string]interface{}{
		"queuekey": key,
		"header":   msg.Header(),
		"body":     msg.Body(),
	})

	uid := strings.ReplaceAll(uuid.New().String(), "-", "")
	newkey := fmt.Sprintf("%s:%s", c.opts.DelayQueueName, uid)

	//过期时间延长的1800,防止服务器时间不一致
	c.client.Set(newkey, string(bytes), time.Second*time.Duration(delaySeconds+int64(c.opts.RangeSeconds)))

	newScore := time.Now().Unix() + delaySeconds
	err := c.client.ZAdd(c.opts.DelayQueueName, &rds.Z{Score: float64(newScore), Member: uid}).Err()
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
	c.onceLock.Do(func() {
		close(c.closeChan)
	})
	return c.client.Close()
}

func (c *Producer) delayQueue() {
	ticker := time.NewTicker(time.Second * time.Duration(c.opts.DelayInterval))
	for {
		select {
		case <-c.closeChan:
			return
		case <-ticker.C:
			c.procDelayQueue(0)
		}
	}
}

func (c *Producer) procDelayQueue(cur int64) {
	vals, err := c.client.ZRangeByScore(c.opts.DelayQueueName, &rds.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(cur, 10),
	}).Result()
	if err != nil {
		log.Errorf("redis.procDelayQueue:%s,err:%+v", c.opts.DelayQueueName, err)
		return
	}
	args := make([]interface{}, len(vals), len(vals))
	for i := range vals {
		args[i] = vals[i]
		c.procDelayItem(vals[i])
	}

	err = c.client.ZRem(c.opts.DelayQueueName, args...).Err()
}

func (c *Producer) procDelayItem(uid string) {
	newkey := fmt.Sprintf("%s:%s", c.opts.DelayQueueName, uid)
	val := c.client.Get(newkey).Val()
	if val == "" {
		return
	}
	msg := newMsgBody(val)
	c.Push(msg.QueueKey, msg)
	c.client.Del(newkey)
}

type producerResolver struct {
}

func (s *producerResolver) Name() string {
	return Proto
}
func (s *producerResolver) Resolve(config config.Config) (queue.IMQP, error) {
	return NewProducer(config)
}
func init() {
	queue.RegisterProducer(&producerResolver{})
}
