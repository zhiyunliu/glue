package queue

import (
	"errors"
	"strings"

	"context"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/session"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var Nil error = errors.New("queue nil")

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
	if len(strings.TrimSpace(key)) == 0 {
		return errors.New("queue.Send,queue name can't be empty")
	}
	msg, err := q.buildMessage(ctx, value)
	if err != nil {
		return err
	}

	msg.Header()[constants.HeaderSourceIp] = global.LocalIp
	msg.Header()[constants.HeaderSourceName] = global.AppName
	return q.q.Push(ctx, key, msg)
}

func (q *queue) BatchSend(ctx context.Context, key string, values ...interface{}) error {
	if len(strings.TrimSpace(key)) == 0 {
		return errors.New("queue.BatchSend,queue name can't be empty")
	}
	msgList := make([]Message, 0, len(values))
	for i := range values {
		msg, err := q.buildMessage(ctx, values[i])
		if err != nil {
			return err
		}
		msgList = append(msgList, msg)
	}

	return q.q.BatchPush(ctx, key, msgList...)
}

func (q *queue) DelaySend(ctx context.Context, key string, value interface{}, delaySeconds int64) error {
	if len(strings.TrimSpace(key)) == 0 {
		return errors.New("queue.DelaySend,queue name can't be empty")
	}
	msg, err := q.buildMessage(ctx, value)
	if err != nil {
		return err
	}
	return q.q.DelayPush(ctx, key, msg, delaySeconds)
}

func (q *queue) buildMessage(ctx context.Context, value any) (msg Message, err error) {
	if value == nil {
		err = errors.New("queue:value can't be null")
		return
	}

	msg, ok := value.(Message)
	if !ok {
		msg = NewMsg(value)
	}
	if sid, ok := session.FromContext(ctx); ok {
		msg.Header()[constants.HeaderRequestId] = sid
	}
	// 注入跟踪信息到请求头
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.MapCarrier(msg.Header()))
	return msg, nil
}

// // Count 队列中消息个数
// func (q *queue) Count(key string) (int64, error) {
// 	return q.q.Count(key)
// }

func (q *queue) Close() error {
	return q.q.Close()
}
