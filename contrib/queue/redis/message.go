package redis

import (
	"encoding/json"

	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/extlib/xtypes"
	"github.com/zhiyunliu/velocity/queue"
)

//RedisMessage reids消息
type redisMessage struct {
	message string
}

//Ack 确定消息
func (m *redisMessage) Ack() error {
	return nil
}

//Nack 取消消息
func (m *redisMessage) Nack() error {
	return nil
}

//GetMessage 获取消息
func (m *redisMessage) GetMessage() queue.Message {
	return newMsgBody(m.message)
}

type MsgBody struct {
	HeaderMap xtypes.SMap `json:"header"`
	BodyMap   xtypes.SMap `json:"body"`
}

func newMsgBody(msg string) *MsgBody {
	body := &MsgBody{}
	json.Unmarshal([]byte(msg), body)
	return body
}

func (m *MsgBody) Header() context.Header {
	return m.HeaderMap
}
func (m *MsgBody) Body() context.Body {
	return m.BodyMap
}
func (m *MsgBody) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
