package alloter

import (
	"bytes"
	sctx "context"
	"encoding/json"
	"io"
	"net/url"
	"strconv"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xmqc"
	"github.com/zhiyunliu/golibs/session"

	"github.com/zhiyunliu/glue/queue"
)

var _ engine.Request = (*Request)(nil)

// Request 处理任务请求
type Request struct {
	ctx  sctx.Context
	task *xmqc.Task
	queue.IMQCMessage
	method string
	url    *url.URL
	params map[string]string
	header engine.Header
	body   cbody
}

// NewRequest 构建任务请求
func newRequest(task *xmqc.Task, m queue.IMQCMessage) (r *Request) {
	r = &Request{
		IMQCMessage: m,
		task:        task,
		method:      string(engine.MethodPost),
		params:      make(map[string]string),
		header:      engine.Header{},
	}

	//将消息原串转换为map
	message := m.GetMessage()
	mheader := message.Header()
	if len(mheader) > 0 {
		for k, v := range mheader {
			r.header.Set(k, v)
		}
	}
	r.body = message.Body()
	r.ctx = sctx.Background()
	r.header.Set("retry_count", strconv.FormatInt(m.RetryCount(), 10))
	r.header.Set("x-xmqc-msg-id", m.MessageId())
	r.header.Set(constants.ContentTypeName, constants.ContentTypeApplicationJSON)

	return r
}

func (m Request) GetSid() string {
	if m.header.Get(constants.HeaderRequestId) == "" {
		m.header.Set(constants.HeaderRequestId, session.Create())
	}
	return m.header.Get(constants.HeaderRequestId)
}

// GetName 获取任务名称
func (m *Request) GetName() string {
	return m.task.Queue
}

// GetService 服务名
func (m *Request) GetService() string {
	return m.task.GetService()
}

// GetService 服务名()
func (m *Request) GetURL() *url.URL {
	if m.url == nil {
		m.url, _ = url.Parse(m.task.GetFullPath())
	}
	return m.url
}

// GetMethod 方法名
func (m *Request) GetMethod() string {
	return m.method
}

func (m *Request) Params() map[string]string {
	return m.params
}

func (m *Request) GetHeader() engine.Header {
	return m.header
}

func (m *Request) Body() []byte {
	return m.body
}

func (m *Request) GetRemoteAddr() string {
	return m.header.Get(constants.HeaderRemoteHeader)
}

func (m *Request) Context() sctx.Context {
	return m.ctx
}
func (m *Request) WithContext(ctx sctx.Context) {
	m.ctx = ctx
}

type Body interface {
	io.Reader
	Scan(obj interface{}) error
}

type cbody []byte

func (b cbody) Read(p []byte) (n int, err error) {
	return bytes.NewReader(b).Read(p)
}

func (b cbody) Scan(obj interface{}) error {
	return json.Unmarshal(b, obj)
}
