package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

// RedisMessage reids消息
type redisMessage struct {
	retryCount int64
	message    string
	obj        *MsgBody
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
	if m.obj == nil {
		m.obj = newMsgBody(m.message)
	}
	return m.obj
}

func (m *redisMessage) PlusRetryCount() queue.Message {
	if m.obj == nil {
		m.obj = newMsgBody(m.message)
	}
	m.retryCount++
	m.obj.HeaderMap["retry_count"] = strconv.FormatInt(m.retryCount, 10)
	return m.obj
}

type MsgBody struct {
	msg       string      `json:"-"`
	QueueKey  string      `json:"qk,omitempty"`
	HeaderMap xtypes.SMap `json:"header,omitempty"`
	BodyMap   xtypes.XMap `json:"body,omitempty"`
}

func newMsgBody(msg string) *MsgBody {
	if !json.Valid(bytesconv.StringToBytes(msg)) {
		panic(fmt.Errorf("msg data is invalid json format.:%s", msg))
	}
	body := &MsgBody{
		msg:       msg,
		HeaderMap: make(xtypes.SMap),
		BodyMap:   make(xtypes.XMap),
	}
	decoder := json.NewDecoder(strings.NewReader(msg))
	decoder.UseNumber()
	decoder.Decode(body)
	return body
}

func (m *MsgBody) Header() map[string]string {
	return m.HeaderMap
}
func (m *MsgBody) Body() map[string]interface{} {
	return m.BodyMap
}

func (m MsgBody) String() string {
	return m.msg
}

func (m MsgBody) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

type deadMsg struct {
	Queue string `json:"q"`
	Msg   string `json:"m"`
}

func (m deadMsg) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
