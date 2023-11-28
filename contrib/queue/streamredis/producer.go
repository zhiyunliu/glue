package streamredis

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
	"github.com/zhiyunliu/golibs/xtypes"
	redisqueue "github.com/zhiyunliu/redisqueue/v2"
)

const (
	DELAY_QUEUE_NAME = "glue:delayqueue:stream"
)

// Producer memcache配置文件
type Producer struct {
	client    *redis.Client
	producer  *redisqueue.Producer
	opts      *ProductOptions
	closeChan chan struct{}
	onceLock  sync.Once
}

// NewProducerByConfig 根据配置文件创建一个redis连接
func NewProducer(config config.Config, opts ...queue.Option) (m *Producer, err error) {

	m = &Producer{
		closeChan: make(chan struct{}),
	}
	m.client, err = getRedisClient(config, opts...)
	if err != nil {
		return
	}
	copts := &ProductOptions{
		DelayQueueName: DELAY_QUEUE_NAME,
		RangeSeconds:   1800,
		DelayInterval:  5,
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
	go m.delayQueue()
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
	return
}

func (c *Producer) procDelayItem(uid string) {
	newkey := fmt.Sprintf("%s:%s", c.opts.DelayQueueName, uid)
	val := c.client.Get(newkey).Val()
	if val == "" {
		return
	}

	decoder := json.NewDecoder(strings.NewReader(val))
	decoder.UseNumber()

	var data xtypes.XMap = make(map[string]interface{})
	decoder.Decode(&data)

	key := data.GetString(queue.QueueKey)
	c.Push(key, newMsgBody(data))

	c.client.Del(newkey)
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
