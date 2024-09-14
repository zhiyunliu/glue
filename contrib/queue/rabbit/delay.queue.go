package rabbit

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/zhiyunliu/glue/queue"
	"golang.org/x/sync/errgroup"
)

func (p *Producer) appendDelay(ctx context.Context, orgQueue string, msg queue.Message, delaySeconds int64) (err error) {

	tmpProcessor, ok := p.delayQueueMap.Load(orgQueue)
	if !ok {
		actual, loaded := p.delayQueueMap.LoadOrStore(orgQueue, p.newProcessor(orgQueue, p.BatchPush))
		if !loaded {
			processor := actual.(queue.DelayProcessor)
			if err = processor.Start(p.closeChan); err != nil {
				return
			}
		}
		tmpProcessor = actual
	}
	return tmpProcessor.(queue.DelayProcessor).AppendMessage(ctx, msg, delaySeconds)
}

func (p *Producer) BatchPush(ctx context.Context, key string, msgList ...queue.Message) error {
	if len(msgList) == 0 {
		return nil
	}
	for i := range msgList {
		if err := p.Push(ctx, key, msgList[i]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Producer) newProcessor(orgQueue string, callback queue.DelayCallback) queue.DelayProcessor {
	return &delayProcess{
		client:     p.client,
		callback:   callback,
		orgQueue:   orgQueue,
		delayQueue: fmt.Sprintf("%s:delay", orgQueue),
		groups:     &errgroup.Group{},
	}
}

type delayProcess struct {
	client     *rabbitClient
	callback   queue.DelayCallback
	orgQueue   string
	delayQueue string
	groups     *errgroup.Group
}

func (p delayProcess) Start(done chan struct{}) (err error) {
	amqpArgs := amqp.Table{
		"x-delayed-type": "direct",
	}

	opts := p.client.options
	channel := p.client.channel
	err = channel.ExchangeDeclare(opts.DelayExchange, "x-delayed-message", true, false, false, false, amqpArgs)
	if err != nil {
		return
	}
	// 声明一个队列
	q, err := channel.QueueDeclare(
		p.orgQueue, // 队列名称
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // 参数
	)
	if err != nil {
		return
	}
	err = channel.QueueBind(
		q.Name,             // queue name
		q.Name,             // routing key
		opts.DelayExchange, // exchange
		false,
		nil,
	)
	return
}

func (p delayProcess) AppendMessage(ctx context.Context, msg queue.Message, delaySeconds int64) (err error) {

	opts := p.client.options

	err = p.client.DelayPublish(ctx, p.orgQueue, opts.DelayExchange, msg, delaySeconds)
	return
}
