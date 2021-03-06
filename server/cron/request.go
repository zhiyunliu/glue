package cron

import (
	"bytes"
	sctx "context"
	"encoding/json"
	"io"
	"time"

	cron "github.com/robfig/cron/v3"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/glue/server"
)

var _ alloter.IRequest = (*Request)(nil)

//Request 处理任务请求
type Request struct {
	ctx      sctx.Context
	job      *Job
	schedule cron.Schedule
	round    *Round
	method   string
	params   map[string]string
	header   map[string]string
	body     cbody //map[string]string
	executed bool
}

//NewRequest 构建任务请求
func NewRequest(job *Job) (r *Request, err error) {

	r = &Request{
		job:    job,
		method: server.MethodGet,
		params: make(map[string]string),
		round:  &Round{},
	}

	r.reset()
	r.body = make(cbody)
	r.schedule, err = cron.ParseStandard(job.Cron)
	if err != nil {
		return r, err
	}

	for k, v := range job.Meta {
		r.body[k] = v
	}

	return r, nil
}

//GetName 服务名
func (m *Request) GetName() string {
	return m.job.Cron
}

//GetService 服务名
func (m *Request) GetService() string {
	return m.job.GetService()
}

//GetMethod 方法名
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
func (m *Request) WithContext(ctx sctx.Context) alloter.IRequest {
	m.ctx = ctx
	return m
}

//NextTime 下次执行时间
func (m *Request) NextTime(t time.Time) time.Time {
	if m.job.IsImmediately() && !m.executed {
		m.executed = true
		return t
	}
	return m.schedule.Next(t)
}
func (m *Request) reset() {
	m.ctx = sctx.Background()
	m.header = make(map[string]string)
	if m.header[constants.ContentTypeName] == "" {
		m.header[constants.ContentTypeName] = constants.ContentTypeApplicationJSON
	}
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
