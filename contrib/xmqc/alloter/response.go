package alloter

import (
	"fmt"
	"net/http"

	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/glue/xmqc"
	"github.com/zhiyunliu/golibs/xtypes"
)

var _ alloter.ResponseWriter = (*Response)(nil)

const (
	noWritten     = -1
	_sucessStatus = http.StatusOK
)

// Request 处理任务请求
type Response struct {
	status int
	size   int
	header xtypes.SMap
	msg    queue.IMQCMessage
	//stream *bufio.Writer
}

// newResponse 构建任务请求
func newResponse(task *xmqc.Task, msg queue.IMQCMessage) (r *Response) {
	r = &Response{
		header: make(xtypes.SMap),
		size:   noWritten,
		status: _sucessStatus,
		msg:    msg,
		//stream: bufio.NewWriter(os.Stdout),
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
	if r.status == _sucessStatus {
		return r.msg.Ack()
	}
	return r.msg.Nack(fmt.Errorf("Status:%d", r.status))
}
