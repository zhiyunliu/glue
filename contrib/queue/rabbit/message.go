package rabbit

import (
	"encoding/json"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xtypes"
)

// rabbitMessage reids消息
type rabbitMessage struct {
	retryCount int64
	message    *amqp.Delivery
	objMsg     queue.Message
	err        error
}

func (m *rabbitMessage) RetryCount() int64 {
	if m.retryCount > 0 {
		return m.retryCount
	}
	rtyCnt := m.GetMessage().Header()["retry_count"]
	if rtyCnt == "" {
		return m.retryCount
	}
	val, _ := strconv.ParseInt(rtyCnt, 10, 32)
	m.retryCount = val
	return m.retryCount
}

func (m *rabbitMessage) MessageId() string {
	return m.message.MessageId
}

// Ack 确定消息
func (m *rabbitMessage) Ack() error {
	m.err = nil
	return nil //	m.message.Ack(false)
}

// Nack 取消消息
func (m *rabbitMessage) Nack(err error) error {
	m.err = err
	return nil //m.message.Nack(false, true)
}

// original message
func (m *rabbitMessage) Original() string {
	return string(m.message.Body)
}

// GetMessage 获取消息
func (m *rabbitMessage) GetMessage() queue.Message {
	if m.objMsg == nil {
		m.objMsg = newMsgBody(m.message.Body)
	}
	return m.objMsg
}

func (m *rabbitMessage) PlusRetryCount() queue.Message {
	if m.objMsg == nil {
		m.objMsg = newMsgBody(m.message.Body)
	}
	m.retryCount++
	m.objMsg.Header()["retry_count"] = strconv.FormatInt(m.retryCount, 10)
	return m.objMsg
}

func newMsgBody(msgBytes []byte) queue.Message {
	msgItem := &queue.MsgItem{
		HeaderMap: make(xtypes.SMap),
	}
	if !json.Valid(msgBytes) {
		return msgItem
	}

	msgItem.ItemBytes = msgBytes
	json.Unmarshal(msgBytes, msgItem)
	return msgItem
}
