package redis

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/zhiyunliu/velocity/components/queues/impls"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/plugins/redis"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zkfy/stompngo"
)

type consumerChan struct {
	msgChan     <-chan stompngo.MessageData
	unconsumeCh chan struct{}
}

//Consumer Consumer
type Consumer struct {
	client  *redis.Client
	queues  cmap.ConcurrentMap
	closeCh chan struct{}
	once    sync.Once
	setting *config.Setting
}

type QueueItem struct {
	QueueName    string
	Concurrency  int //等于0 ，代表不限制
	BlockTimeout int

	onceLock      sync.Once
	unconsumeChan chan struct{}
	consumer      *Consumer
	callback      impls.ConsumeCallback
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
func NewConsumer(setting *config.Setting) (consumer *Consumer, err error) {
	consumer = &Consumer{}
	consumer.setting = setting

	consumer.closeCh = make(chan struct{})
	consumer.queues = cmap.New()
	return
}

//Connect  连接服务器
func (consumer *Consumer) Connect() (err error) {
	consumer.client, err = redis.NewByConfig(consumer.setting)
	return
}

//Consume 注册消费信息
func (consumer *Consumer) Consume(queue string, callback impls.ConsumeCallback) (err error) {
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
	consumer := item.Consumer
	unconsumeChan := item.UnconsumeChan
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
			if err != nil {
				log.Printf("BLPop.%s,Error:%+v", queue, err)
				time.Sleep(time.Second)
				continue
			}
			hasData := len(msgs) > 0
			if !hasData {
				continue
			}
			ndata := msgs[len(msgs)-1]
			go callback(&RedisMessage{Message: ndata, HasData: hasData})
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

func (s *consumeResolver) Resolve(setting *config.Setting) (impls.IMQC, error) {
	return NewConsumer(setting)
}
func init() {
	impls.RegisterConsumer(&consumeResolver{})
}
