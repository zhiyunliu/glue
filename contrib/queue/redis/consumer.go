package redis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	rds "github.com/go-redis/redis"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/contrib/redis"
	"github.com/zhiyunliu/gel/queue"

	cmap "github.com/orcaman/concurrent-map"
)

//Consumer Consumer
type Consumer struct {
	client  *redis.Client
	queues  cmap.ConcurrentMap
	closeCh chan struct{}
	once    sync.Once
	config  config.Config
}

type QueueItem struct {
	QueueName    string
	Concurrency  int //等于0 ，代表不限制
	BlockTimeout int

	onceLock      sync.Once
	unconsumeChan chan struct{}
	consumer      *Consumer
	callback      queue.ConsumeCallback
}

func (item *QueueItem) ReceiveStart() {
	go doReceiveMsg(item)
}

func (item *QueueItem) ReceiveStop() {
	item.onceLock.Do(func() {
		close(item.unconsumeChan)
	})
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
	consumer.client, err = redis.NewByConfig(consumer.config)
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
	item := &QueueItem{
		QueueName:    queue,
		BlockTimeout: 2,

		unconsumeChan: make(chan struct{}),
		consumer:      consumer,
		callback:      callback,
	}

	success := consumer.queues.SetIfAbsent(queue, item)
	if success {
		item.ReceiveStart()
	}
	return
}

func doReceiveMsg(item *QueueItem) {
	consumer := item.consumer
	unconsumeChan := item.unconsumeChan
	client := consumer.client
	queue := item.QueueName
	callback := item.callback

	for {
		select {
		case <-consumer.closeCh:
			return
		case <-unconsumeChan:
			return
		default:
			cmd := client.BLPop(time.Duration(item.BlockTimeout)*time.Second, queue)
			msgs, err := cmd.Result()
			if err != nil && err != rds.Nil {
				time.Sleep(time.Second)
				continue
			}
			hasData := len(msgs) > 0
			if !hasData {
				continue
			}
			ndata := msgs[len(msgs)-1]
			go callback(&redisMessage{message: ndata})
		}
	}
}

//UnConsume 取消注册消费
func (consumer *Consumer) Unconsume(queue string) {
	if consumer.client == nil {
		return
	}
	if item, ok := consumer.queues.Get(queue); ok {
		item.(*QueueItem).ReceiveStop()
	}
	consumer.queues.Remove(queue)
}

func (consumer *Consumer) Start() {

}

//Close 关闭当前连接
func (consumer *Consumer) Close() {
	consumer.once.Do(func() {
		close(consumer.closeCh)
	})

	for item := range consumer.queues.IterBuffered() {
		item.Val.(*QueueItem).ReceiveStop()
	}

	if consumer.client == nil {
		return
	}
	consumer.client.Close()

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
