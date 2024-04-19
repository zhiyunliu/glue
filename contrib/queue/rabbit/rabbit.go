package rabbit

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xnet"
	"golang.org/x/sync/errgroup"
)

var (
	clientCache    = sync.Map{}
	mutex          = sync.Mutex{}
	canRunCheckKey = "rabbit-canrun"
)

type rabbitClient struct {
	consumeTag     string
	ctx            context.Context
	cancelCallback context.CancelFunc
	conn           *amqp.Connection
	channel        *amqp.Channel
	options        *options
	queueCache     sync.Map
	mutex          sync.Mutex
	connNotify     chan *amqp.Error
	channelNotify  chan *amqp.Error
	canRunCheck    sync.Map
}

func defaultOptions() *options {
	return &options{
		ConnName:      "glue-conn",
		Exchange:      "glue-exchange",
		DelayExchange: "glue-delay-exchange",
		ExchangeType:  ExchangeTypeDirect,
		Options:       *queue.DefaultOptions(),
	}
}

func getRabbitClient(config config.Config, opts ...queue.Option) (client *rabbitClient, err error) {

	addr := config.Value("addr").String()

	if tmpClient, ok := clientCache.Load(addr); ok {
		return tmpClient.(*rabbitClient), nil
	}
	mutex.Lock()
	defer mutex.Unlock()

	if tmpClient, ok := clientCache.Load(addr); ok {
		return tmpClient.(*rabbitClient), nil
	}

	protoType, configName, err := xnet.Parse(addr)
	if err != nil {
		panic(err)
	}
	rootCfg := config.Root()
	queueCfg := rootCfg.Get(protoType).Get(configName)

	queueOpts := &queue.Options{}
	for i := range opts {
		opts[i](queueOpts)
	}

	clientCtx, callback := context.WithCancel(context.Background())

	client = &rabbitClient{
		consumeTag:     global.AppName,
		ctx:            clientCtx,
		cancelCallback: callback,
		options:        defaultOptions(),
		canRunCheck:    sync.Map{},
	}

	err = queueCfg.ScanTo(client.options)
	if err != nil {
		err = fmt.Errorf("getRabbitClient.ScanTo:%+v", err)
		return
	}

	for i := range opts {
		opts[i](&client.options.Options)
	}

	err = client.Connect()
	if err != nil {
		return nil, err
	}

	clientCache.Store(addr, client)

	go client.watchConn()
	return
}

func (c *rabbitClient) Connect() (err error) {

	rabbitCfg := amqp.Config{
		Vhost:      c.options.VirtualHost,
		Properties: amqp.NewConnectionProperties(),
	}
	rabbitCfg.Properties.SetClientConnectionName(c.options.ConnName)

	if c.conn == nil || c.conn.IsClosed() {
		c.conn, err = amqp.DialConfig(c.options.Addr, rabbitCfg)
		if err != nil {
			return fmt.Errorf("dial: %s", err)
		}
		c.connNotify = c.conn.NotifyClose(make(chan *amqp.Error))
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("channel: %s", err)
	}
	c.channelNotify = c.channel.NotifyClose(make(chan *amqp.Error))
	return nil
}

func (c *rabbitClient) watchConn() {

	for {
		select {
		case err := <-c.connNotify:
			if err != nil {
				log.Error("watchConn rabbitmq - connection NotifyClose: ", err)
			}
		case err := <-c.channelNotify:
			if err != nil {
				log.Error("watchConn rabbitmq - channel NotifyClose: ", err)
			}
		case <-c.ctx.Done():
			return
		}

		c.reconnect()
	}
}

func (c *rabbitClient) reconnect() {

	canRunLock := make(chan struct{})
	c.canRunCheck.Store(canRunCheckKey, canRunLock)
	defer func() {
		close(canRunLock)
		c.canRunCheck.Delete(canRunCheckKey)
	}()

	if c.channel.IsClosed() {
		for err := range c.channelNotify {
			log.Error("watchConn rabbitmq channelNotify:", err)
		}
	}

	if c.conn.IsClosed() {
		for err := range c.connNotify {
			log.Error("watchConn rabbitmq connNotify:", err)
		}
	}

	for err := c.Connect(); err != nil; {
		log.Error("watchConn rabbitmq Connect:", err)
		time.Sleep(time.Second) //1秒后重试
	}
	log.Error("watchConn rabbitmq ReConnect success")
}

func (c *rabbitClient) getAvalChannel() (channel *amqp.Channel) {
	locker, ok := c.canRunCheck.Load(canRunCheckKey)
	if ok {
		<-locker.(chan struct{})
	}

	for c.conn.IsClosed() {
		log.Error("getAvalChannel rabbitmq conn.IsClosed")
		time.Sleep(time.Second)
	}

	for c.channel.IsClosed() {
		log.Error("getAvalChannel rabbitmq channel.IsClosed")
		time.Sleep(time.Second)
	}
	return c.channel
}

func (c *rabbitClient) Consume(queueName string, meta metadata.Metadata) (delivery chan *amqp.Delivery, err error) {
	if err = c.QueueDeclare(queueName); err != nil {
		return
	}

	delivery = make(chan *amqp.Delivery)
	group := errgroup.Group{}
	group.Go(func() error {
		for {
			channel := c.getAvalChannel()
			curDelivery, err := channel.Consume(queueName, c.consumeTag, true, false, false, false, amqp.Table(meta))
			if err != nil {
				err = fmt.Errorf("channel.Consume:%+v,err:%+v", queueName, err)
				log.Error("rabbitmq consume :", err)
				time.Sleep(time.Second) //1s后重试
				continue
			}

			for item := range curDelivery {
				delivery <- &item
			}

			for len(delivery) > 0 {
				time.Sleep(time.Second)
			}
		}
	})

	return
}

func (c *rabbitClient) ExchangeDeclare() (err error) {
	err = c.channel.ExchangeDeclare(c.options.Exchange, c.options.ExchangeType, false, false, false, false, nil)
	if err != nil {
		err = fmt.Errorf("channel.ExchangeDeclare:%+v", err)
		return
	}
	return
}

func (c *rabbitClient) QueueDeclare(queueName string) (err error) {
	_, ok := c.queueCache.Load(queueName)
	if ok {
		return
	}
	c.mutex.Lock()
	defer func() {
		if err == nil {
			c.queueCache.Store(queueName, true)
		} else {
			c.queueCache.Delete(queueName)
		}
		c.mutex.Unlock()
	}()

	if _, ok := c.queueCache.Load(queueName); ok {
		return
	}

	_, err = c.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		err = fmt.Errorf("channel.QueueDeclare:%+v,err:%+v", queueName, err)
		return
	}

	if err = c.channel.QueueBind(
		queueName,
		queueName,
		c.options.Exchange, // sourceExchange
		false,              // noWait
		nil,                // arguments
	); err != nil {
		return fmt.Errorf("queue Bind: %s,err:%+v", queueName, err)
	}
	return
}

func (c *rabbitClient) DelayPublish(queueName, exchange string, msg queue.Message, delaySeconds int64) (err error) {

	amqpMsg := amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Persistent,
		Headers: amqp.Table{
			"x-delay": delaySeconds * 1000,
		},
	}
	amqpMsg.Body, err = msg.MarshalBinary()
	if err != nil {
		err = fmt.Errorf("publish.MarshalBinary:%+v", err)
		return
	}

	channel := c.getAvalChannel()
	err = channel.PublishWithContext(context.Background(), exchange, queueName, false, false, amqpMsg)
	if err != nil {
		err = fmt.Errorf("DelayPublish.PublishWithContext:%+v", err)
		return
	}
	return nil
}

func (c *rabbitClient) Publish(queueName string, msg queue.Message) (err error) {
	if err = c.QueueDeclare(queueName); err != nil {
		return
	}
	amqpMsg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
	}
	amqpMsg.Body, err = msg.MarshalBinary()
	if err != nil {
		err = fmt.Errorf("publish.MarshalBinary:%+v", err)
		return
	}
	channel := c.getAvalChannel()
	err = channel.PublishWithContext(context.Background(), c.options.Exchange, queueName, false, false, amqpMsg)
	if err != nil {
		err = fmt.Errorf("publish.PublishWithContext:%+v", err)
		return
	}
	return
}

func (c *rabbitClient) Close() error {
	c.cancelCallback()
	if !c.channel.IsClosed() {
		c.channel.Close()
	}
	if !c.conn.IsClosed() {
		c.conn.Close()
	}
	return nil
}
