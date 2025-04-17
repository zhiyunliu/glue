package robfigcron

import (
	"bytes"
	sctx "context"
	"encoding/json"
	"net/url"
	"sync"
	"sync/atomic"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xcron"
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xtypes"
)

var _ engine.Request = (*Request)(nil)

// Request 处理任务请求
type Request struct {
	ctx     sctx.Context
	job     *xcron.Job
	method  string
	url     *url.URL
	params  xtypes.SMap
	header  engine.Header
	body    *cbody //map[string]string
	session string
	canProc uint32
	mu      sync.Mutex
}

// NewRequest 构建任务请求
func newRequest(job *xcron.Job) (r *Request) {
	r = &Request{
		job:    job,
		method: string(engine.MethodPost),
		params: make(map[string]string),
	}

	r.reset()
	r.body = &cbody{
		data: make(map[string]interface{}),
	}

	for k, v := range job.Meta {
		r.body.data[k] = v
	}
	return r
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

func (m *Request) GetHeader() engine.Header {

	return m.header
}

func (m *Request) Body() []byte {
	return m.body.Bytes()
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

func (m *Request) CanProc() bool {
	//0 false,1 true
	oldv := atomic.LoadUint32(&m.canProc)
	if oldv == 1 && atomic.CompareAndSwapUint32(&m.canProc, oldv, 0) {
		return true
	}
	return false

	// if m.canProc {
	// 	m.canProc = false
	// 	return true
	// }
	// return false
}

func (m *Request) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreUint32(&m.canProc, 1)
	m.session = session.Create()
	m.header = make(map[string]string)
	m.header.Set(constants.ContentTypeName, constants.ContentTypeApplicationJSON)
	m.header.Set(constants.HeaderRequestId, m.session)
	m.header.Set("x-cron-engine", Proto)
	m.header.Set("x-cron-job-key", m.job.GetKey())
}

func (m *Request) Monopoly(monopolyJobs cmap.ConcurrentMap[string, *monopolyJob]) (bool, error) {
	//本身不是独占
	if !m.job.IsMonopoly() {
		return false, nil
	}

	mjob, ok := monopolyJobs.Get(m.job.GetKey())
	//独占列表不存在（只存在close的短暂时间）
	if !ok {
		return true, nil
	}

	isSuc, err := mjob.Acquire()
	if err != nil {
		return true, err
	}
	if isSuc {
		return false, nil
	}
	return true, nil
}

type cbody struct {
	reader *bytes.Reader
	bytes  []byte
	data   map[string]interface{}
}

func (b *cbody) Bytes() []byte {
	if b.bytes == nil {
		b.bytes, _ = json.Marshal(b.data)
		b.reader = bytes.NewReader(b.bytes)
	}
	return b.bytes
}

func (b *cbody) Read(p []byte) (n int, err error) {
	return b.reader.Read(p)
}
