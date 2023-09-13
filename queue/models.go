package queue

import (
	"bytes"
	"encoding/json"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

type MsgWrap struct {
	HeaderMap xtypes.SMap `json:"header,omitempty"`
	BodyMap   xtypes.XMap `json:"body"`
	hasProced bool        `json:"-"`
	reqid     string      `json:"-"`
	bodyBytes []byte      `json:"-"`
}

type Option func(m *MsgWrap)

func WithXRequestID(reqId string) Option {
	return func(m *MsgWrap) {
		m.reqid = reqId
	}
}

func WithHeader(key, val string) Option {
	return func(m *MsgWrap) {
		if m.HeaderMap == nil {
			m.HeaderMap = make(xtypes.SMap)
		}
		m.HeaderMap[key] = val
	}
}

func NewMsg(obj interface{}, opts ...Option) (msg Message, err error) {
	var bytes []byte
	switch val := obj.(type) {
	case []byte:
		bytes = val
	case Message:
		return val, nil
	default:
		bytes, err = json.Marshal(obj)
		if err != nil {
			return nil, err
		}
	}
	tmpmsg := &MsgWrap{
		bodyBytes: bytes,
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
	if w.HeaderMap == nil {
		w.HeaderMap = map[string]string{}
	}
	w.BodyMap = map[string]interface{}{}
	w.hasProced = true
	if w.reqid != "" {
		w.HeaderMap[constants.HeaderRequestId] = w.reqid
	}
	decoder := json.NewDecoder(bytes.NewReader(w.bodyBytes))
	decoder.UseNumber()
	decoder.Decode(&w.BodyMap)
}

func (w *MsgWrap) String() string {
	return bytesconv.BytesToString(w.bodyBytes)
}
