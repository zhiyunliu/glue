package xipc

import (
	"fmt"
	"net/http"

	"github.com/zhiyunliu/golibs/httputil"

	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/golibs/bytesconv"
)

var (
	_ Body           = (*Response)(nil)
	_ ResponseEntity = (*Response)(nil)
)

type ResponseEntity = httputil.ResponseEntity
type Body = httputil.Body

// 定义请求消息
type Request struct {
	Service string        `json:"service,omitempty"` // 服务名称
	Method  string        `json:"method,omitempty"`  // 请求方法, GET
	Header  engine.Header `json:"header,omitempty"`  // 请求头信息 map[string]string
	Body    []byte        `json:"body,omitempty"`    // 请求参数 json
}

func (x *Request) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *Request) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Request) GetHeader() map[string]string {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Request) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

// 定义响应消息
type Response struct {
	Status  int32             `json:"status,omitempty"` // 状态码
	Headers map[string]string `json:"header,omitempty"` // 返回头信息，map[string]string
	Result  []byte            `json:"result,omitempty"` // 返回结果，default=json,according to content-type
}

func (b *Response) GetStatus() int32 {
	return b.Status
}
func (b *Response) GetHeader() map[string]string {
	return b.Headers
}
func (b *Response) GetResult() []byte {
	return b.Result
}

func (b *Response) StatusCode() int {
	return int(b.Status)
}
func (b *Response) Header() map[string]string {
	return b.Headers
}
func (b *Response) Body() (bytes []byte, err error) {
	return b.Result, nil
}

func BadRequest(err error) *Response {
	return &Response{
		Status: int32(http.StatusBadRequest),
		Result: bytesconv.StringToBytes(fmt.Sprintf("请求帧有误:%v", err)),
	}
}
