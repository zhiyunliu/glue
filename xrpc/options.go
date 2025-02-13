package xrpc

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/xtypes"
)

type RequestOption func(*Options)

type Options struct {
	Header       xtypes.SMap // 请求头
	Method       string      // 请求方法
	Query        string      // 请求参数
	WaitForReady bool        // 是否等待服务端响应
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
		o.Header = header
	}
}

// WithXRequestID 设置请求ID
func WithXRequestID(requestID string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.HeaderRequestId] = requestID
	}
}

// WithWaitForReady 设置是否等待服务端响应
func WithWaitForReady(waitForReady bool) RequestOption {
	return func(o *Options) {
		o.WaitForReady = waitForReady
	}
}

// WithContentType 设置请求内容类型
func WithContentType(contentType string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.ContentTypeName] = contentType
	}
}

// WithSourceName 设置来源服务名
func WithSourceName() RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.HeaderSourceName] = global.AppName
	}
}
