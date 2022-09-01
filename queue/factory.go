package queue

import (
	"encoding/json"
	"errors"
	"fmt"

	"context"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/golibs/xtypes"
)

var Nil error = errors.New("Queue Nil")

//queue 对输入KEY进行封装处理
type queue struct {
	q IMQP
}

func newQueue(proto string, cfg config.Config) (IQueue, error) {
	var err error
	q := &queue{}
	q.q, err = NewMQP(proto, cfg)
	return q, err
}

//Send 发送消息
func (q *queue) Send(ctx context.Context, key string, value interface{}) error {
	if msg, ok := value.(Message); ok {
		return q.q.Push(key, msg)
	}

	msg, err := newMsgWrap("", value)
	if err != nil {
		return fmt.Errorf("queue.Send:%s,Error:%w", key, err)
	}

	return q.q.Push(key, msg)
}
func (q *queue) DelaySend(ctx context.Context, key string, value interface{}, delaySeconds int64) error {
	if msg, ok := value.(Message); ok {
		return q.q.Push(key, msg)
	}

	msg, err := newMsgWrap("", value)
	if err != nil {
		return fmt.Errorf("queue.Send:%s,Error:%w", key, err)
	}

	return q.q.DelayPush(key, msg, delaySeconds)
}

// //Pop 从队列中获取一个消息
// func (q *queue) Pop(key string) (string, error) {
// 	return q.q.Pop(key)
// }

//Count 队列中消息个数
func (q *queue) Count(key string) (int64, error) {
	return q.q.Count(key)
}

func (q *queue) Close() error {
	return q.q.Close()
}

type msgWrap struct {
	HeaderMap xtypes.SMap `json:"header"`
	BodyMap   xtypes.XMap `json:"body"`
	hasProced bool        `json:"-"`
	reqid     string      `json:"-"`
	bytes     []byte      `json:"-"`
}

func newMsgWrap(reqid string, obj interface{}) (msg Message, err error) {
	bytes, ok := obj.([]byte)
	if !ok {
		bytes, err = json.Marshal(obj)
		if err != nil {
			return nil, err
		}
	}

	return &msgWrap{
		reqid: reqid,
		bytes: bytes,
	}, nil
}

func (w *msgWrap) Header() map[string]string {
	w.adapterMsg()
	return w.HeaderMap
}
func (w *msgWrap) Body() map[string]interface{} {
	w.adapterMsg()
	return w.BodyMap

}

func (w *msgWrap) adapterMsg() {
	if w.hasProced {
		return
	}
	w.HeaderMap = map[string]string{}
	w.BodyMap = map[string]interface{}{}
	w.hasProced = true
	json.Unmarshal(w.bytes, &w.BodyMap)
}

func (w *msgWrap) String() string {
	return string(w.bytes)
}
