package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// 响应状态码。
const (
	CodeOK             = 0
	CodeBadRequest     = 400
	CodeUnknownCommand = 404
	CodeHandlerError   = 500
	CodeHandlerPanic   = 501
	CodeEncodeError    = 502
)

// Request 是 Handler 收到的统一请求视图。
type Request interface {
	Command() string
	RawData() json.RawMessage
	Bind(v any) error
}

type jsonRequest struct {
	cmd  string
	data json.RawMessage
}

func (r *jsonRequest) Command() string { return r.cmd }

func (r *jsonRequest) RawData() json.RawMessage {
	return r.data
}

func (r *jsonRequest) Bind(v any) error {
	if len(r.data) == 0 {
		return validateBindTarget(v)
	}
	return json.Unmarshal(r.data, v)
}

// Response 是服务端返回给客户端的统一响应。
type Response struct {
	Code    int    `json:"code"`
	SubCode string `json:"sub_code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    any    `json:"data,omitempty"`

	rawData json.RawMessage
}

// IsOK 判断响应是否成功。
func (r *Response) IsOK() bool {
	return r != nil && r.Code == CodeOK
}

// Bind 将响应数据解析到 v 指向的结构。
func (r *Response) Bind(v any) error {
	if r == nil {
		return errors.New("ipc: cannot bind a nil response")
	}
	if err := validateBindTarget(v); err != nil {
		return err
	}
	if len(r.rawData) == 0 {
		if r.Data == nil {
			return nil
		}
		data, err := json.Marshal(r.Data)
		if err != nil {
			return fmt.Errorf("ipc: encode response data: %w", err)
		}
		return json.Unmarshal(data, v)
	}
	return json.Unmarshal(r.rawData, v)
}

func validateBindTarget(v any) error {
	value := reflect.ValueOf(v)
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() {
		return errors.New("ipc: bind target must be a non-nil pointer")
	}
	return nil
}

// UnmarshalJSON 解析响应并保留原始 data，以便后续进行强类型绑定。
func (r *Response) UnmarshalJSON(data []byte) error {
	var wire struct {
		Code    int             `json:"code"`
		SubCode string          `json:"sub_code"`
		Msg     string          `json:"msg"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	r.Code = wire.Code
	r.SubCode = wire.SubCode
	r.Msg = wire.Msg
	r.rawData = append(r.rawData[:0], wire.Data...)
	if len(wire.Data) == 0 {
		r.Data = nil
		return nil
	}
	return json.Unmarshal(wire.Data, &r.Data)
}

// OK 创建成功响应。
func OK(data any) *Response {
	return &Response{Code: CodeOK, Data: data}
}

// Fail 创建失败响应。
func Fail(code int, subCode, msg string) *Response {
	return &Response{Code: code, SubCode: subCode, Msg: msg}
}

// FailWithData 创建包含附加数据的失败响应。
func FailWithData(code int, subCode, msg string, data any) *Response {
	return &Response{Code: code, SubCode: subCode, Msg: msg, Data: data}
}

// WrapError 将 error 转换为统一失败响应。
func WrapError(code int, subCode string, err error) *Response {
	if err == nil {
		return OK(nil)
	}
	return Fail(code, subCode, err.Error())
}

// Handler 处理一个 IPC 命令请求。
type Handler interface {
	Handle(req Request) *Response
}

// HandlerFunc 让普通函数可以直接注册为命令处理器。
type HandlerFunc func(req Request) *Response

func (f HandlerFunc) Handle(req Request) *Response {
	return f(req)
}
