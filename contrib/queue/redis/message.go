package redis

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

// RedisMessage reids消息
type redisMessage struct {
	retryCount int64
	message    string
	objMsg     queue.Message
	err        error
}

func (m *redisMessage) RetryCount() int64 {
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

func (m *redisMessage) MessageId() string {
	return ""
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
	return m.message
}

// GetMessage 获取消息
func (m *redisMessage) GetMessage() queue.Message {
	if m.objMsg == nil {
		m.objMsg = newMsgBody(m.message)
	}
	return m.objMsg
}

func (m *redisMessage) PlusRetryCount() queue.Message {
	if m.objMsg == nil {
		m.objMsg = newMsgBody(m.message)
	}

	newMsg := &queue.MsgItem{
		HeaderMap: m.objMsg.Header(),
		BodyBytes: m.objMsg.Body(),
	}

	m.retryCount++
	newMsg.HeaderMap["retry_count"] = strconv.FormatInt(m.retryCount, 10)
	return newMsg
}

//{"user_id":123}
//{"header":{},"body":{"user_id":123}}

func newMsgBody(msg string) queue.Message {
	msgBytes := bytesconv.StringToBytes(msg)
	if !json.Valid(msgBytes) {
		panic(fmt.Errorf("msg data is invalid json format.:%s", msg))
	}
	msgItem := &queue.MsgItem{
		HeaderMap: make(xtypes.SMap),
		BodyBytes: msgBytes,
	}
	// msgItem.ItemBytes = msgBytes
	json.Unmarshal(msgBytes, msgItem)
	return msgItem
}

type deadMsg struct {
	Queue string `json:"q"`
	Msg   string `json:"m"`
}

func (m deadMsg) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
