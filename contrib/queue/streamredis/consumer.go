package streamredis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xtypes"

	redisqueue "github.com/zhiyunliu/redisqueue/v2"

	cmap "github.com/orcaman/concurrent-map"
)

// Consumer Consumer
type Consumer struct {
	configName       string
	EnableDeadLetter bool //开启死信队列
	DeadLetterQueue  string
	queues           cmap.ConcurrentMap
	consumer         *redisqueue.Consumer
	producer         *redisqueue.Producer
	redisClient      *redis.Client
	closeCh          chan struct{}

	once   sync.Once
	config config.Config
}

type QueueItem struct {
	QueueName         string
	Concurrency       int
	BufferSize        int
	VisibilityTimeout time.Duration
	callback          queue.ConsumeCallback
}

func (q QueueItem) GetQueue() string {
	return q.QueueName
}

func (q QueueItem) GetConcurrency() int {
	return q.Concurrency
}

func (s QueueItem) GetVisibilityTimeout() time.Duration {
	return s.VisibilityTimeout
}

func (s QueueItem) GetBufferSize() int {
	return s.BufferSize
}

// NewConsumerByConfig 创建新的Consumer
func NewConsumer(configName string, config config.Config) (consumer *Consumer, err error) {
	consumer = &Consumer{}
	consumer.configName = configName
	consumer.config = config
	consumer.closeCh = make(chan struct{})
	consumer.queues = cmap.New()
	return
}

// Connect  连接服务器
func (consumer *Consumer) Connect() (err error) {
	client, err := getRedisClient(consumer.config)
	if err != nil {
		return
	}
	copts := &ConsumerOptions{}
	err = consumer.config.ScanTo(copts)
	if err != nil {
		return
	}
	consumer.EnableDeadLetter = len(copts.DeadLetterQueue) > 0
	consumer.DeadLetterQueue = copts.DeadLetterQueue
	consumer.redisClient = client
	opts := &redisqueue.ConsumerOptions{
		GroupName:         global.AppName,
		RedisClient:       client.UniversalClient,
		Concurrency:       100,
		BufferSize:        1000,
		BlockingTimeout:   2 * time.Second,
		VisibilityTimeout: 30 * time.Second,
		ReclaimInterval:   5 * time.Second,
	}
	if len(copts.GroupName) > 0 {
		opts.GroupName = copts.GroupName
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
	if copts.VisibilityTimeout > 0 {
		opts.VisibilityTimeout = time.Duration(copts.VisibilityTimeout) * time.Second
	}
	if copts.ReclaimInterval > 0 {
		opts.ReclaimInterval = time.Duration(copts.ReclaimInterval) * time.Second
	}
	err = consumer.createProducer()
	if err != nil {
		return
	}
	consumer.consumer, err = redisqueue.NewConsumerWithOptions(opts)
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case <-consumer.closeCh:
				return
			case err := <-consumer.consumer.Errors:
				log.Error(err)
				continue
			}
		}
	}()
	return
}

func (m *Consumer) createProducer() (err error) {
	copts := &ProductOptions{
		DelayInterval: 2,
	}
	err = m.config.ScanTo(copts)
	if err != nil {
		return
	}

	pdtOpts := &redisqueue.ProducerOptions{
		StreamMaxLength:      10000,
		RedisClient:          m.redisClient.UniversalClient,
		ApproximateMaxLength: copts.ApproximateMaxLength,
	}
	if copts.StreamMaxLength > 0 {
		pdtOpts.StreamMaxLength = copts.StreamMaxLength
	}

	m.producer, err = redisqueue.NewProducerWithOptions(pdtOpts)
	return
}

// Consume 注册消费信息
func (consumer *Consumer) Consume(task queue.TaskInfo, callback queue.ConsumeCallback) (err error) {
	queueName := task.GetQueue()
	if strings.EqualFold(queueName, "") {
		return fmt.Errorf("队列名字不能为空")
	}
	if callback == nil {
		return fmt.Errorf("queue:%s,回调函数不能为nil", queueName)
	}

	item := &QueueItem{
		QueueName:         queueName,
		Concurrency:       task.GetConcurrency(),
		BufferSize:        task.GetBufferSize(),
		VisibilityTimeout: time.Duration(task.GetVisibilityTimeout()) * time.Second,
		callback:          callback,
	}
	if item.Concurrency == 0 {
		item.Concurrency = queue.DefaultMaxQueueLen
	}

	consumer.queues.SetIfAbsent(queueName, item)

	return
}

// UnConsume 取消注册消费
func (consumer *Consumer) Unconsume(queue string) {
	consumer.queues.Remove(queue)
}

func (consumer *Consumer) Start() error {
	for item := range consumer.queues.IterBuffered() {
		tqi := item.Val.(*QueueItem)
		var confunc redisqueue.ConsumerFunc = func(qi *QueueItem) redisqueue.ConsumerFunc {
			return func(m *redisqueue.Message) error {
				if m.RetryCount >= queue.MaxRetrtCount {
					//todo:写入死信队列
					consumer.writeToDeadLetter(tqi.QueueName, m.Values)
					return nil
				}
				msg := &redisMessage{message: m.Values, retryCount: m.RetryCount, messageId: m.ID}
				qi.callback(msg)
				return msg.Error()
			}
		}(tqi)
		consumer.consumer.Register(tqi, confunc)
	}

	go consumer.consumer.Run()
	return nil
}

// Close 关闭当前连接
func (consumer *Consumer) Close() error {
	consumer.once.Do(func() {
		close(consumer.closeCh)
	})

	consumer.consumer.Shutdown()
	return nil
}

func (consumer *Consumer) writeToDeadLetter(queue string, vals xtypes.XMap) {
	if !consumer.EnableDeadLetter {
		return
	}
	if strings.EqualFold(queue, consumer.DeadLetterQueue) {
		return
	}
	deadMsg := make(xtypes.XMap)

	deadMsg["q"] = queue
	deadMsg["m"] = vals

	consumer.producer.Enqueue(context.Background(), &redisqueue.Message{Stream: consumer.DeadLetterQueue, Values: deadMsg})
}

type consumeResolver struct {
}

func (s *consumeResolver) Name() string {
	return Proto
}

func (s *consumeResolver) Resolve(configName string, setting config.Config) (queue.IMQC, error) {
	return NewConsumer(configName, setting)
}
func init() {
	queue.RegisterConsumer(&consumeResolver{})
}
