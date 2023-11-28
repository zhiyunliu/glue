package redis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	rds "github.com/go-redis/redis/v7"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/queue"

	cmap "github.com/orcaman/concurrent-map"
)

// Consumer Consumer
type Consumer struct {
	configName       string
	EnableDeadLetter bool //开启死信队列
	DeadLetterQueue  string
	client           *redis.Client
	queues           cmap.ConcurrentMap
	closeCh          chan struct{}
	once             sync.Once
	wg               sync.WaitGroup
	config           config.Config
}

type QueueItem struct {
	QueueName    string
	Concurrency  int //等于0 ，代表不限制
	BlockTimeout int

	closeMsgChanLock *sync.Once
	unconsumeChan    chan struct{}
	msgChan          chan string
	callback         queue.ConsumeCallback
}

func (qi *QueueItem) CloseMsgChan() {
	qi.closeMsgChanLock.Do(func() {
		close(qi.msgChan)
	})
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
	consumer.client, err = getRedisClient(consumer.config)
	consumer.DeadLetterQueue = consumer.config.Value("deadletter_queue").String()
	consumer.EnableDeadLetter = len(consumer.DeadLetterQueue) > 0
	return
}

// Consume 注册消费信息
func (consumer *Consumer) Consume(task queue.TaskInfo, callback queue.ConsumeCallback) (err error) {
	queue := task.GetQueue()
	if strings.EqualFold(queue, "") {
		return fmt.Errorf("队列名字不能为空")
	}
	if callback == nil {
		return fmt.Errorf("queue:%s,回调函数不能为nil", queue)
	}
	item := &QueueItem{
		QueueName:        queue,
		Concurrency:      task.GetConcurrency(),
		BlockTimeout:     2,
		unconsumeChan:    make(chan struct{}),
		callback:         callback,
		closeMsgChanLock: &sync.Once{},
	}

	consumer.queues.SetIfAbsent(queue, item)
	return
}

func (consumer *Consumer) doReceive(item *QueueItem) {
	client := consumer.client
	queueName := item.QueueName
	concurrency := item.Concurrency
	if concurrency == 0 {
		concurrency = queue.DefaultMaxQueueLen //无限制时候，默认500个，如果消息没有正常处理，最多造成1000个消息丢失
	}
	item.Concurrency = concurrency
	item.msgChan = make(chan string, concurrency)

	consumer.wg.Add(concurrency)

	for i := 0; i < item.Concurrency; i++ {
		go consumer.work(item)
	}

	for {
		select {
		case <-consumer.closeCh:
			close(item.msgChan)
			return
		case <-item.unconsumeChan:
			close(item.msgChan)
			return
		default:
			cmd := client.BLPop(time.Duration(item.BlockTimeout)*time.Second, queueName)
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
			item.msgChan <- ndata
		}
	}
}

func (consumer *Consumer) stopReceive(item *QueueItem) {
	close(item.unconsumeChan)
}

func (consumer *Consumer) work(item *QueueItem) {
	defer func() {
		for data := range item.msgChan {
			//回填消息队列数据
			consumer.client.LPush(item.QueueName, data)
		}
		consumer.wg.Done()
	}()
	for {
		select {
		case msg := <-item.msgChan:
			rdsMsg := &redisMessage{message: msg}
			item.callback(rdsMsg)
			if rdsMsg.err != nil {
				//超过最大次数
				if rdsMsg.RetryCount() >= queue.MaxRetrtCount {
					consumer.writeToDeadLetter(item.QueueName, msg)
					continue
				}
				obj := rdsMsg.PlusRetryCount()
				consumer.client.RPush(item.QueueName, obj)
			}
		case <-consumer.closeCh:
			return
		case <-item.unconsumeChan:
			return
		}
	}
}

// UnConsume 取消注册消费
func (consumer *Consumer) Unconsume(queue string) {
	if consumer.client == nil {
		return
	}
	if item, ok := consumer.queues.Get(queue); ok {
		consumer.stopReceive(item.(*QueueItem))
	}
	consumer.queues.Remove(queue)
}

// Start 启动
func (consumer *Consumer) Start() {
	for item := range consumer.queues.IterBuffered() {
		func(qitem *QueueItem) {
			go consumer.doReceive(qitem)
		}(item.Val.(*QueueItem))
	}
}

// Close 关闭当前连接
func (consumer *Consumer) Close() {
	consumer.once.Do(func() {
		close(consumer.closeCh)
	})
	//等等所有的关闭完成
	consumer.wg.Wait()
	if consumer.client == nil {
		return
	}
	consumer.client.Close()
}

func (consumer *Consumer) writeToDeadLetter(queue string, msg string) {
	if !consumer.EnableDeadLetter {
		return
	}

	if strings.EqualFold(queue, consumer.DeadLetterQueue) {
		return
	}
	consumer.client.RPush(consumer.DeadLetterQueue, deadMsg{Queue: queue, Msg: msg})
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
