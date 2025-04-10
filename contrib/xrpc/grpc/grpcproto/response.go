package grpcproto

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xrpc"
)

var (
	_ xrpc.Body             = (*Response)(nil)
	_ engine.ResponseEntity = (*Response)(nil)
)

// Deprecated: As of Go v0.5.22, this function simply calls [StatusCode].
func (x *Response) GetStatus() int32 {
	return int32(x.StatusCode())
}

// Deprecated: As of Go v0.5.22, this function simply calls [Header].
func (x *Response) GetHeader() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

// Deprecated: As of Go v0.5.22, this function simply calls [Body].
func (x *Response) GetResult() (body []byte) {
	body, _ = x.Body()
	return
}

func (x *Response) StatusCode() int {
	if x != nil {
		return int(x.Status)
	}
	return 0
}

func (x *Response) Header() map[string]string {
	if x != nil {
		return x.Headers
	}
	return map[string]string{}
}

func (x *Response) Body() (bytes []byte, err error) {
	if x != nil {
		return x.Result, nil
	}
	return []byte{}, nil
}
