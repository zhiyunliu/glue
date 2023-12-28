package alloter

import (
	"bytes"
	sctx "context"
	"encoding/json"
	"io"
	"net/url"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xcron"
	"github.com/zhiyunliu/golibs/session"
)

var _ engine.Request = (*Request)(nil)

// Request 处理任务请求
type Request struct {
	ctx          sctx.Context
	job          *xcron.Job
	round        *Round
	url          *url.URL
	method       string
	params       map[string]string
	header       map[string]string
	body         cbody //map[string]string
	session      string
	canProc      bool
	CalcNextTime time.Time
}

// NewRequest 构建任务请求
func newRequest(job *xcron.Job) (r *Request, err error) {

	r = &Request{
		job:    job,
		method: engine.MethodPost,
		params: make(map[string]string),
		round:  &Round{Job: job},
	}

	r.reset()
	r.body = make(cbody)
	if err != nil {
		return r, err
	}

	for k, v := range job.Meta {
		r.body[k] = v
	}

	return r, nil
}

// GetName 服务名
func (m *Request) GetName() string {
	return m.job.Cron
}

// GetService 服务名
func (m *Request) GetService() string {
	return m.job.GetService()
}

// GetURL 服务URL
func (m *Request) GetURL() *url.URL {
	if m.url == nil {
		m.url, _ = url.Parse(m.job.GetService())
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

func (m *Request) GetHeader() map[string]string {
	return m.header
}

func (m *Request) Body() []byte {
	bytes, _ := json.Marshal(m.body)
	return bytes
}

func (m *Request) GetRemoteAddr() string {
	return m.header[constants.HeaderRemoteHeader]
}

func (m *Request) Context() sctx.Context {
	return m.ctx
}
func (m *Request) WithContext(ctx sctx.Context) {
	m.ctx = ctx
}

func (m *Request) CanProc() bool {
	if m.canProc {
		m.canProc = false
		return true
	}
	return false
}

func (m *Request) reset() {
	m.canProc = true
	m.session = session.Create()
	//m.ctx = sctx.Background()
	m.header = make(map[string]string)
	m.header[constants.ContentTypeName] = constants.ContentTypeApplicationJSON
	m.header[constants.HeaderRequestId] = m.session
	m.header["x-cron-engine"] = Proto
}

func (m *Request) Monopoly(monopolyJobs cmap.ConcurrentMap) (bool, error) {
	//本身不是独占
	if !m.job.IsMonopoly() {
		return false, nil
	}

	val, ok := monopolyJobs.Get(m.job.GetKey())
	//独占列表不存在（只存在close的短暂时间）
	if !ok {
		return true, nil
	}
	mjob := val.(*monopolyJob)
	isSuc, err := mjob.Acquire()
	if err != nil {
		return true, err
	}
	if isSuc {
		return false, nil
	}
	return true, nil
}

type Body interface {
	io.Reader
	Scan(obj interface{}) error
}

type cbody map[string]interface{}

func (b cbody) Read(p []byte) (n int, err error) {
	bodyBytes, err := json.Marshal(b)
	if err != nil {
		return 0, err
	}
	return bytes.NewReader(bodyBytes).Read(p)
}

func (b cbody) Scan(obj interface{}) error {
	bytes, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, obj)
}
