package rabbit

import (
	"fmt"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/queue"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// Consumer Consumer
type Consumer struct {
	configName       string
	EnableDeadLetter bool //开启死信队列
	DeadLetterQueue  string
	client           *rabbitClient
	queues           cmap.ConcurrentMap[string, *QueueItem]
	closeCh          chan struct{}
	once             sync.Once
	wg               sync.WaitGroup
	config           config.Config
}

type QueueItem struct {
	QueueName        string
	taskInfo         queue.TaskInfo
	closeMsgChanLock *sync.Once
	unconsumeChan    chan struct{}
	msgChan          chan *amqp.Delivery
	callback         queue.ConsumeCallback
}

// NewConsumerByConfig 创建新的Consumer
func NewConsumer(configName string, config config.Config) (consumer *Consumer, err error) {
	consumer = &Consumer{}
	consumer.configName = configName
	consumer.config = config

	consumer.closeCh = make(chan struct{})
	consumer.queues = cmap.New[*QueueItem]()

	consumer.client, err = getRabbitClient(config)
	if err != nil {
		return consumer, err
	}
	return
}

// Connect  连接服务器
func (c *Consumer) Connect() (err error) {

	err = c.client.ExchangeDeclare()
	if err != nil {
		return err
	}

	c.DeadLetterQueue = c.config.Value("deadletter_queue").String()
	c.EnableDeadLetter = len(c.DeadLetterQueue) > 0
	return
}

// Consume 注册消费信息
func (consumer *Consumer) Consume(taskInfo queue.TaskInfo, callback queue.ConsumeCallback) (err error) {
	queue := taskInfo.GetQueue()
	if strings.EqualFold(queue, "") {
		return fmt.Errorf("队列名字不能为空")
	}
	if callback == nil {
		return fmt.Errorf("queue:%s,回调函数不能为nil", queue)
	}
	item := &QueueItem{
		QueueName:        queue,
		taskInfo:         taskInfo,
		unconsumeChan:    make(chan struct{}),
		callback:         callback,
		closeMsgChanLock: &sync.Once{},
	}

	consumer.queues.SetIfAbsent(queue, item)
	return
}

func (consumer *Consumer) doReceive(item *QueueItem) {

	concurrency := item.taskInfo.GetConcurrency()
	if concurrency == 0 {
		concurrency = queue.DefaultMaxQueueLen
	}
	item.msgChan = make(chan *amqp.Delivery, concurrency)

	consumer.wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go consumer.work(item)
	}

	msgChan, err := consumer.client.Consume(item.QueueName, item.taskInfo.GetMeta())
	if err != nil {
		return
	}

	for {
		select {
		case <-consumer.closeCh:
			close(item.msgChan)
			return
		case <-item.unconsumeChan:
			close(item.msgChan)
			return
		case msgItem := <-msgChan:
			item.msgChan <- msgItem
		}
	}
}

func (consumer *Consumer) stopReceive(item *QueueItem) {
	close(item.unconsumeChan)
}

func (consumer *Consumer) work(item *QueueItem) {
	defer consumer.wg.Done()
	for msg := range item.msgChan {
		rdsMsg := &rabbitMessage{message: msg}
		item.callback(rdsMsg)
		if rdsMsg.err != nil {
			//超过最大次数
			if rdsMsg.RetryCount() >= queue.MaxRetrtCount {
				consumer.writeToDeadLetter(item.QueueName, msg)
				continue
			}
		}
	}
}

// UnConsume 取消注册消费
func (consumer *Consumer) Unconsume(queue string) {
	if consumer.client == nil {
		return
	}
	if item, ok := consumer.queues.Get(queue); ok {
		consumer.stopReceive(item)
	}
	consumer.queues.Remove(queue)
}

// Start 启动
func (consumer *Consumer) Start() (err error) {
	for item := range consumer.queues.IterBuffered() {
		func(qitem *QueueItem) {
			go consumer.doReceive(qitem)
		}(item.Val)
	}
	return nil
}

// Close 关闭当前连接
func (c *Consumer) Close() (err error) {
	c.once.Do(func() {
		c.client.Close()
		close(c.closeCh)
	})
	//等等所有的关闭完成
	c.wg.Wait()
	return
}

func (consumer *Consumer) writeToDeadLetter(queue string, msg *amqp.Delivery) {
	if !consumer.EnableDeadLetter {
		return
	}

	if strings.EqualFold(queue, consumer.DeadLetterQueue) {
		return
	}
	//todo:处理死信内容的写入
	//consumer.client.RPush(consumer.DeadLetterQueue, deadMsg{Queue: queue, Msg: msg})
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
