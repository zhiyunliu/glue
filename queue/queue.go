package queue

import (
	"context"
	"encoding/json"

	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

//默认最大队列长度 500
var DefaultMaxQueueLen = 100

//IQueue 消息队列
type IQueue interface {
	Send(ctx context.Context, key string, value interface{}) error
	DelaySend(ctx context.Context, key string, value interface{}, delaySeconds int64) error
	Count(key string) (int64, error)
}

//IMQCMessage  队列消息
type IMQCMessage interface {
	RetryCount() int64
	Ack() error
	Nack(error) error
	Original() string
	GetMessage() Message
}

type TaskInfo interface {
	GetQueue() string
	GetConcurrency() int
}

type Message interface {
	Header() map[string]string
	Body() map[string]interface{}
	String() string
}

type ConsumeCallback func(IMQCMessage)

//IMQC consumer接口
type IMQC interface {
	Connect() error
	Consume(task TaskInfo, callback ConsumeCallback) (err error)
	Unconsume(queue string)
	Start()
	Close()
}

//IMQP 消息生产
type IMQP interface {
	Push(key string, value Message) error
	DelayPush(key string, value Message, delaySeconds int64) error
	Count(key string) (int64, error)
	Close() error
}

//IComponentQueue Component Queue
type IComponentQueue interface {
	GetQueue(name string) (q IQueue)
}

type MsgItem struct {
	HeaderMap xtypes.SMap `json:"header"`
	BodyMap   xtypes.XMap `json:"body"`
	strval    string      `json:"-"`
}

func (w *MsgItem) Header() map[string]string {
	return w.HeaderMap
}

func (w *MsgItem) Body() map[string]interface{} {
	return w.BodyMap
}

func (w *MsgItem) String() string {
	if w.strval == "" {
		bytes, _ := json.Marshal(w)
		w.strval = bytesconv.BytesToString(bytes)
	}
	return w.strval
}
