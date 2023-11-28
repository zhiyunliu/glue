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

const (
	DELAY_QUEUE_NAME = "glue:delayqueue:list"
)

// Producer memcache配置文件
type Producer struct {
	opts      *ProductOptions
	client    *redis.Client
	closeChan chan struct{}
	onceLock  sync.Once
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config, opts ...queue.Option) (m *Producer, err error) {
	m = &Producer{}
	m.client, err = getRedisClient(config, opts...)
	if err != nil {
		return
	}
	m.opts = &ProductOptions{
		DelayQueueName: DELAY_QUEUE_NAME,
		RangeSeconds:   1800,
		DelayInterval:  5,
	}
	err = config.Scan(m.opts)
	if err != nil {
		return
	}
	go m.delayQueue()
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
		queue.QueueKey: key,
		"header":       msg.Header(),
		"body":         msg.Body(),
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
		case now := <-ticker.C:
			c.procDelayQueue(now.Unix())
		}
	}
}

func (p *Producer) procDelayQueue(cur int64) {
	vals, err := p.client.ZRangeByScore(p.opts.DelayQueueName, &rds.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(cur, 10),
	}).Result()
	if err != nil {
		log.Errorf("streamredis.procDelayQueue.ZRangeByScore:%s,err:%+v", p.opts.DelayQueueName, err)
		return
	}
	if len(vals) == 0 {
		return
	}
	//每次处理的命令条数
	const CMD_COUNT = 100
	tmpLen := len(vals)
	if tmpLen > CMD_COUNT {
		tmpLen = CMD_COUNT
	}

	cycCnt := len(vals) / tmpLen
	if cycCnt*tmpLen < len(vals) {
		cycCnt = cycCnt + 1
	}

	idx := 0
	totalLen := len(vals)
	isLast := false
	for c := 0; c < cycCnt; c++ {
		args := make([]interface{}, 0, tmpLen)
		cycIdx := c * tmpLen
		isLast = c == (cycCnt - 1)
		for i := 0; i < tmpLen; i++ {
			idx = cycIdx + i
			args = append(args, vals[idx])
			p.procDelayItem(vals[idx])
			if isLast && (idx+1) == totalLen {
				break
			}
		}
		err = p.client.ZRem(p.opts.DelayQueueName, args...).Err()
		if err != nil {
			log.Errorf("streamredis.procDelayQueue.ZRem:%s,err:%+v", p.opts.DelayQueueName, err)
		}
	}

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
func (s *producerResolver) Resolve(config config.Config, opts ...queue.Option) (queue.IMQP, error) {
	return NewProducer(config, opts...)
}
func init() {
	queue.RegisterProducer(&producerResolver{})
}
