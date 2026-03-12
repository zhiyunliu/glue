package robfigcron

import (
	"net/http"

	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/golibs/engine"
)

var _ alloter.ResponseWriter = (*Response)(nil)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// Request 处理任务请求
type Response struct {
	status int
	size   int
	header engine.Header
	//stream *bufio.Writer
}

// newResponse 构建任务请求
func newResponse() (r *Response) {
	r = &Response{
		header: make(engine.Header),
		size:   noWritten,
		status: defaultStatus,
	}
	return r
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
func (r *Response) Header() engine.Header {
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
