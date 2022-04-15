package rpc

import (
	"bufio"
	"bytes"
	"net/http"

	"github.com/zhiyunliu/gel/contrib/alloter"
	"github.com/zhiyunliu/golibs/xtypes"
)

var _ alloter.ResponseWriter = (*Response)(nil)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

//Request 处理任务请求
type Response struct {
	status    int
	size      int
	header    xtypes.SMap
	dataBytes []byte
	data      *bufio.Writer

	//stream *bufio.Writer
}

//NewRequest 构建任务请求
func NewResponse() (r *Response) {

	r = &Response{
		header:    make(xtypes.SMap),
		size:      noWritten,
		status:    defaultStatus,
		dataBytes: make([]byte, 0),
	}
	buffer := bytes.NewBuffer(r.dataBytes)
	r.data = bufio.NewWriter(buffer)
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
	return r.data.Write(data)
}

// Writes the string into the response body.
func (r *Response) WriteString(s string) (n int, err error) {
	r.size += n
	return r.data.WriteString(s)
}

func (r *Response) Flush() {

}
