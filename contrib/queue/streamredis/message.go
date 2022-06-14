package streamredis

import (
	"encoding/json"

	"github.com/zhiyunliu/gel/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

//RedisMessage reids消息
type redisMessage struct {
	message map[string]interface{}
	err     error
	obj     *MsgBody
}

func (m *redisMessage) Error() error {
	return m.err
}

//Ack 确定消息
func (m *redisMessage) Ack() error {
	m.err = nil
	return nil
}

//Nack 取消消息
func (m *redisMessage) Nack(err error) error {
	m.err = err
	return nil
}

//original message
func (m *redisMessage) Original() string {
	if m.obj == nil {
		m.obj = newMsgBody(m.message)
	}
	return m.obj.String()
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
	HeaderMap xtypes.SMap `json:"header"`
	BodyMap   xtypes.XMap `json:"body"`
}

func newMsgBody(msg map[string]interface{}) *MsgBody {

	msgBytes, _ := json.Marshal(msg)
	body := &MsgBody{
		msg:       bytesconv.BytesToString(msgBytes),
		HeaderMap: make(xtypes.SMap),
		BodyMap:   make(xtypes.XMap),
	}
	json.Unmarshal(msgBytes, body)
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
