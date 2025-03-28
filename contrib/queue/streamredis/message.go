package streamredis

import (
	"encoding/json"
	"fmt"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

// RedisMessage reids消息
type redisMessage struct {
	retryCount int64
	messageId  string
	message    map[string]interface{}
	err        error
	obj        *queue.MsgItem
}

func (m redisMessage) MessageId() string {
	return m.messageId
}

func (m redisMessage) Error() error {
	return m.err
}

func (m redisMessage) RetryCount() int64 {
	return m.retryCount
}

// Ack 确定消息
func (m *redisMessage) Ack() error {
	m.err = nil
	return nil
}

// Nack 取消消息
func (m *redisMessage) Nack(err error) error {
	m.err = err
	return nil
}

// original message
func (m *redisMessage) Original() string {
	if m.obj == nil {
		m.obj = newMsgBody(m.message)
	}
	return m.obj.String()
}

// GetMessage 获取消息
func (m *redisMessage) GetMessage() queue.Message {
	if m.obj == nil {
		m.obj = newMsgBody(m.message)
	}
	return m.obj
}

// type MsgBody struct {
// 	msg       []byte      `json:"-"`
// 	HeaderMap xtypes.SMap `json:"header"`
// 	BodyMap   xtypes.XMap `json:"body"`
// }

func newMsgBody(msg map[string]interface{}) *queue.MsgItem {
	body := &queue.MsgItem{
		HeaderMap: make(xtypes.SMap),
	}
	switch val := msg["header"].(type) {
	case string:
		_ = json.Unmarshal([]byte(val), &body.HeaderMap)
	case map[string]interface{}:
		for k, v := range val {
			body.HeaderMap[k] = fmt.Sprint(v)
		}
	default:

	}

	switch val := msg["body"].(type) {
	case []byte:
		body.BodyBytes = val
	case string:
		body.BodyBytes = bytesconv.StringToBytes(val)
	default:

	}

	body.ItemBytes, _ = json.Marshal(body)
	return body
}
