package queue

import (
	"encoding/json"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

type MsgWrap struct {
	HeaderMap xtypes.SMap     `json:"header,omitempty"`
	BodyBytes json.RawMessage `json:"body"`
	hasProced bool            `json:"-"`
	reqid     string          `json:"-"`
}

type MsgOption func(m *MsgWrap)

func WithXRequestID(reqId string) MsgOption {
	return func(m *MsgWrap) {
		m.reqid = reqId
	}
}

func WithHeader(key, val string) MsgOption {
	return func(m *MsgWrap) {
		if m.HeaderMap == nil {
			m.HeaderMap = make(xtypes.SMap)
		}
		m.HeaderMap[key] = val
	}
}

func NewMsg(obj interface{}, opts ...MsgOption) (msg Message) {
	var bytes []byte
	switch val := obj.(type) {
	case []byte:
		bytes = val
	case string:
		bytes = bytesconv.StringToBytes(val)
	case Message:
		return val
	default:
		bytes, _ = json.Marshal(obj)
	}
	tmpmsg := &MsgWrap{
		HeaderMap: make(xtypes.SMap),
		BodyBytes: bytes,
	}
	for i := range opts {
		opts[i](tmpmsg)
	}

	return tmpmsg
}

func (w *MsgWrap) Header() map[string]string {
	w.adapterMsg()
	return w.HeaderMap
}
func (w *MsgWrap) Body() []byte {
	w.adapterMsg()
	return w.BodyBytes

}

func (w MsgWrap) String() string {
	return bytesconv.BytesToString(w.BodyBytes)
}

func (w MsgWrap) MarshalBinary() (data []byte, err error) {
	return json.Marshal(w)
}

func (w *MsgWrap) adapterMsg() {
	if w.hasProced {
		return
	}
	if w.HeaderMap == nil {
		w.HeaderMap = map[string]string{}
	}
	w.hasProced = true
	if w.reqid != "" {
		w.HeaderMap[constants.HeaderRequestId] = w.reqid
	}
	w.HeaderMap[constants.HeaderSourceIp] = global.LocalIp
}
