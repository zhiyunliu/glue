package queue

import (
	"errors"
	"fmt"

	"context"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/global"
)

var Nil error = errors.New("Queue Nil")

// queue 对输入KEY进行封装处理
type queue struct {
	q IMQP
}

func newQueue(proto string, cfg config.Config, opts ...Option) (IQueue, error) {
	var err error
	q := &queue{}
	q.q, err = NewMQP(proto, cfg, opts...)
	return q, err
}

// Send 发送消息
func (q *queue) Send(ctx context.Context, key string, value interface{}) error {
	if msg, ok := value.(Message); ok {
		return q.q.Push(key, msg)
	}

	msg, err := NewMsg(value)
	if err != nil {
		return fmt.Errorf("queue.Send:%s,Error:%w", key, err)
	}

	msg.Header()[constants.HeaderSourceIp] = global.LocalIp
	return q.q.Push(key, msg)
}
func (q *queue) DelaySend(ctx context.Context, key string, value interface{}, delaySeconds int64) error {
	if msg, ok := value.(Message); ok {
		return q.q.DelayPush(key, msg, delaySeconds)
	}

	msg, err := NewMsg(value)
	if err != nil {
		return fmt.Errorf("queue.Send:%s,Error:%w", key, err)
	}
	msg.Header()[constants.HeaderSourceIp] = global.LocalIp
	return q.q.DelayPush(key, msg, delaySeconds)
}

// //Pop 从队列中获取一个消息
// func (q *queue) Pop(key string) (string, error) {
// 	return q.q.Pop(key)
// }

// Count 队列中消息个数
func (q *queue) Count(key string) (int64, error) {
	return q.q.Count(key)
}

func (q *queue) Close() error {
	return q.q.Close()
}
