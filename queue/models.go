package queue

import (
	"bytes"
	"encoding/json"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/golibs/xtypes"
)

type MsgWrap struct {
	HeaderMap xtypes.SMap `json:"header,omitempty"`
	BodyMap   xtypes.XMap `json:"body"`
	hasProced bool        `json:"-"`
	reqid     string      `json:"-"`
	bytes     []byte      `json:"-"`
}

type Option func(m *MsgWrap)

func WithXRequestID(reqId string) Option {
	return func(m *MsgWrap) {
		m.reqid = reqId
	}
}

func NewMsg(obj interface{}, opts ...Option) (msg Message, err error) {
	bytes, ok := obj.([]byte)
	if !ok {
		bytes, err = json.Marshal(obj)
		if err != nil {
			return nil, err
		}
	}
	tmpmsg := &MsgWrap{
		bytes: bytes,
	}
	for i := range opts {
		opts[i](tmpmsg)
	}

	return tmpmsg, nil
}

func (w *MsgWrap) Header() map[string]string {
	w.adapterMsg()
	return w.HeaderMap
}
func (w *MsgWrap) Body() map[string]interface{} {
	w.adapterMsg()
	return w.BodyMap

}

func (w *MsgWrap) adapterMsg() {
	if w.hasProced {
		return
	}
	w.HeaderMap = map[string]string{}
	w.BodyMap = map[string]interface{}{}
	w.hasProced = true
	w.HeaderMap[constants.HeaderRequestId] = w.reqid
	decoder := json.NewDecoder(bytes.NewReader(w.bytes))
	decoder.UseNumber()
	decoder.Decode(&w.BodyMap)
}

func (w *MsgWrap) String() string {
	return string(w.bytes)
}
