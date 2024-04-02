package rabbit

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xnet"
)

type rabbitClient struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	options    *options
	queueCache sync.Map
	mutex      sync.Mutex
}

func defaultOptions() *options {
	return &options{
		ConnName:     "glue-conn",
		Exchange:     "glue-exchange",
		ExchangeType: ExchangeTypeDirect,
		Options:      *queue.DefaultOptions(),
	}
}

func getRabbitClient(config config.Config, opts ...queue.Option) (client *rabbitClient, err error) {

	addr := config.Value("addr").String()
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

	client = &rabbitClient{
		options: defaultOptions(),
	}

	err = queueCfg.ScanTo(client.options)
	if err != nil {
		err = fmt.Errorf("getRabbitClient.ScanTo:%+v", err)
		return
	}

	for i := range opts {
		opts[i](&client.options.Options)
	}

	rabbitCfg := amqp.Config{
		Vhost:      client.options.VirtualHost,
		Properties: amqp.NewConnectionProperties(),
	}
	rabbitCfg.Properties.SetClientConnectionName(client.options.ConnName)

	client.conn, err = amqp.DialConfig(client.options.Addr, rabbitCfg)
	if err != nil {
		return client, fmt.Errorf("dial: %s", err)
	}
	client.channel, err = client.conn.Channel()
	if err != nil {
		return client, fmt.Errorf("channel: %s", err)
	}
	return
}

func (c *rabbitClient) Close() error {
	//todo
	c.channel.Close()
	c.conn.Close()
	return nil
}

func (c *rabbitClient) Consume(queueName, appName string, meta metadata.Metadata) (delivery <-chan amqp.Delivery, err error) {
	if err = c.QueueDeclare(queueName); err != nil {
		return
	}

	delivery, err = c.channel.Consume(queueName, appName, false, false, false, false, amqp.Table(meta))
	if err != nil {
		err = fmt.Errorf("channel.Consume:%+v,err:%+v", queueName, err)
		return
	}
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
	err = c.channel.PublishWithContext(context.Background(), c.options.Exchange, queueName, false, false, amqpMsg)
	if err != nil {
		err = fmt.Errorf("publish.PublishWithContext:%+v", err)
		return
	}
	return
}
