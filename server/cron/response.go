package cron

import (
	"net/http"

	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/golibs/xtypes"
)

var _ alloter.ResponseWriter = (*Response)(nil)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

//Request 处理任务请求
type Response struct {
	status int
	size   int
	header xtypes.SMap
	//stream *bufio.Writer
}

//NewRequest 构建任务请求
func NewResponse(job *Job) (r *Response, err error) {
	r = &Response{
		header: make(xtypes.SMap),
		size:   noWritten,
		status: defaultStatus,
	}
	return r, nil
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int {
	return r.size
}

// Returns true if the response body was already written.
func (r *Response) Written() bool {
	return r.size != noWritten

}

func (r *Response) WriteHeader(code int) {
	if code > 0 && r.status != code {
		r.status = code
	}
}
func (r *Response) Header() xtypes.SMap {
	return r.header
}
func (r *Response) Write(data []byte) (n int, err error) {
	r.size += len(data)
	return
}

// Writes the string into the response body.
func (r *Response) WriteString(s string) (n int, err error) {
	r.size += n
	return
}

func (r *Response) Flush() error {
	return nil
}
