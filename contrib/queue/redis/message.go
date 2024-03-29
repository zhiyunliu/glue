package redis

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

//RedisMessage reids消息
type redisMessage struct {
	message string
	obj     *MsgBody
}

func (m *redisMessage) RetryCount() int64 {
	return 0
}

//Ack 确定消息
func (m *redisMessage) Ack() error {
	return nil
}

//Nack 取消消息
func (m *redisMessage) Nack(error) error {
	return nil
}

//original message
func (m *redisMessage) Original() string {
	return m.message
}

//GetMessage 获取消息
func (m *redisMessage) GetMessage() queue.Message {
	if m.obj == nil {
		m.obj = newMsgBody(m.message)
	}
	return m.obj
}

type MsgBody struct {
	msg       string      `json:"-"`
	QueueKey  string      `json:"queuekey"`
	HeaderMap xtypes.SMap `json:"header"`
	BodyMap   xtypes.XMap `json:"body"`
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

func (m *MsgBody) String() string {
	return m.msg
}
