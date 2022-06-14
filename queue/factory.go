package queue

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/constants"
	"github.com/zhiyunliu/gel/context"
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

	reqid := ctx.Header(constants.HeaderRequestId)

	if msg, ok := value.(Message); ok {
		header := msg.Header()
		header[constants.HeaderRequestId] = reqid
		return q.q.Push(key, msg)
	}

	msg, err := newMsgWrap(reqid, value)
	if err != nil {
		return fmt.Errorf("queue.Send:%s,Error:%w", key, err)
	}

	return q.q.Push(key, msg)
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
	if w.HeaderMap == nil {
		w.HeaderMap = map[string]string{
			constants.HeaderRequestId: w.reqid,
		}
	}
	return w.HeaderMap
}
func (w *msgWrap) Body() map[string]interface{} {
	if w.BodyMap == nil {
		w.BodyMap = map[string]interface{}{}
		json.Unmarshal(w.bytes, &w.BodyMap)
	}
	return w.BodyMap

}

func (w *msgWrap) String() string {
	return string(w.bytes)
}
