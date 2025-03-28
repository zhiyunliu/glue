package queue

import (
	"context"
	"encoding"
	"encoding/json"

	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

// 默认最大队列长度 100
var DefaultMaxQueueLen = 100

// IQueue 消息队列
type IQueue interface {
	//发送消息
	Send(ctx context.Context, key string, value any) error
	//批量发送消息
	BatchSend(ctx context.Context, key string, value ...any) error
	//延迟发送消息
	DelaySend(ctx context.Context, key string, value any, delaySeconds int64) error
}

// IMQCMessage  队列消息
type IMQCMessage interface {
	MessageId() string
	RetryCount() int64
	Ack() error
	Nack(error) error
	Original() string
	GetMessage() Message
}

type TaskInfo interface {
	GetQueue() string
	GetConcurrency() int
	GetVisibilityTimeout() int
	GetBufferSize() int
	GetMeta() metadata.Metadata
}

type Message interface {
	encoding.BinaryMarshaler
	Header() map[string]string
	Body() []byte
	String() string
}

type ConsumeCallback func(IMQCMessage)

// IMQC consumer接口
type IMQC interface {
	ServerURL() string
	Connect() error
	Consume(task TaskInfo, callback ConsumeCallback) (err error)
	Unconsume(queue string)
	Start() error
	Close() error
}

// IMQP 消息生产
type IMQP interface {
	Push(ctx context.Context, key string, value Message) error
	BatchPush(ctx context.Context, key string, value ...Message) error
	DelayPush(ctx context.Context, key string, value Message, delaySeconds int64) error
	Close() error
}

// IComponentQueue Component Queue
type IComponentQueue interface {
	GetQueue(name string) (q IQueue)
}

type MsgItem struct {
	HeaderMap xtypes.SMap     `json:"header"`
	BodyBytes json.RawMessage `json:"body"`
	ItemBytes json.RawMessage `json:"-"`
}

func (w MsgItem) MarshalBinary() (data []byte, err error) {
	if len(w.ItemBytes) > 0 {
		return w.ItemBytes, nil
	}
	return json.Marshal(w)
}

func (w MsgItem) Header() map[string]string {
	return w.HeaderMap
}

func (w MsgItem) Body() []byte {
	return w.BodyBytes
}

func (w *MsgItem) String() string {
	if len(w.ItemBytes) == 0 {
		w.ItemBytes, _ = json.Marshal(w)
	}
	return bytesconv.BytesToString(w.ItemBytes)
}
