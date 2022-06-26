package streamredis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/queue"

	redisqueue "github.com/robinjoseph08/redisqueue/v2"

	cmap "github.com/orcaman/concurrent-map"
)

//Consumer Consumer
type Consumer struct {
	queues   cmap.ConcurrentMap
	consumer *redisqueue.Consumer
	closeCh  chan struct{}

	once   sync.Once
	config config.Config
}

type QueueItem struct {
}

//NewConsumerByConfig 创建新的Consumer
func NewConsumer(config config.Config) (consumer *Consumer, err error) {
	consumer = &Consumer{}
	consumer.config = config

	consumer.closeCh = make(chan struct{})
	consumer.queues = cmap.New()
	return
}

//Connect  连接服务器
func (consumer *Consumer) Connect() (err error) {
	client, err := getRedisClient(consumer.config)
	if err != nil {
		return
	}
	copts := &ConsumerOptions{}
	err = consumer.config.Scan(copts)
	if err != nil {
		return
	}

	opts := &redisqueue.ConsumerOptions{
		RedisClient: client.UniversalClient,
	}
	if copts.Concurrency > 0 {
		opts.Concurrency = copts.Concurrency
	}
	if copts.BufferSize > 0 {
		opts.BufferSize = copts.BufferSize
	}
	if copts.BlockingTimeout > 0 {
		opts.BlockingTimeout = time.Duration(copts.BlockingTimeout) * time.Second
	}

	consumer.consumer, err = redisqueue.NewConsumerWithOptions(opts)
	if err != nil {
		return
	}

	return
}

//Consume 注册消费信息
func (consumer *Consumer) Consume(queue string, callback queue.ConsumeCallback) (err error) {
	if strings.EqualFold(queue, "") {
		return fmt.Errorf("队列名字不能为空")
	}
	if callback == nil {
		return fmt.Errorf("queue:%s,回调函数不能为nil", queue)
	}
	item := &QueueItem{}
	success := consumer.queues.SetIfAbsent(queue, item)
	if success {
		consumer.consumer.Register(queue, func(m *redisqueue.Message) error {
			msg := &redisMessage{message: m.Values}
			callback(msg)
			return msg.Error()
		})
	}
	return
}

//UnConsume 取消注册消费
func (consumer *Consumer) Unconsume(queue string) {
	consumer.queues.Remove(queue)
}

func (consumer *Consumer) Start() {
	go consumer.consumer.Run()
}

//Close 关闭当前连接
func (consumer *Consumer) Close() {
	consumer.once.Do(func() {
		close(consumer.closeCh)
	})

	consumer.consumer.Shutdown()
}

type consumeResolver struct {
}

func (s *consumeResolver) Name() string {
	return Proto
}

func (s *consumeResolver) Resolve(setting config.Config) (queue.IMQC, error) {
	return NewConsumer(setting)
}
func init() {
	queue.RegisterConsumer(&consumeResolver{})
}
