package xipc

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
)

type RequestOption func(*Options)

type Options struct {
	Header             engine.Header // 请求头
	Method             string        // 请求方法
	Query              string        // 请求参数
	WaitForReady       bool          // 是否等待服务端响应
	MaxCallRecvMsgSize int           // 最大接收消息体大小，默认4M(maximum message size in bytes the client can receive)
	MaxCallSendMsgSize int           // 最大发送消息体大小，默认4M(maximum message size in bytes the client can send)
	StreamProcessor    any           // 是否使用流传输
}

// WithQuery 设置请求参数
func WithQuery(query string) RequestOption {
	return func(o *Options) {
		o.Query = query
	}
}

// WithMethod 设置请求方法
func WithMethod(method string) RequestOption {
	return func(o *Options) {
		if method != "" {
			o.Method = method
		}
	}
}

// WithHeaders 设置请求头
func WithHeaders(header map[string]string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		for k, v := range header {
			o.Header.Set(k, v)
		}
	}
}

// WithXRequestID 设置请求ID
func WithXRequestID(requestID string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header.Set(constants.HeaderRequestId, requestID)
	}
}

// WithContentType 设置请求内容类型
func WithContentType(contentType string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header.Set(constants.ContentTypeName, contentType)
	}
}

// WithSourceName 设置来源服务名
func WithSourceName() RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header.Set(constants.HeaderSourceName, global.AppName)
	}
}
