package xhttp

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/httputil"
	"github.com/zhiyunliu/golibs/xsse"
	"github.com/zhiyunliu/golibs/xtypes"
)

type RequestOption func(*Options)
type RespHandler = httputil.RespHandler
type SSEHandler = httputil.SSEHandler
type SSEOption = httputil.SSEOption
type ReqChangeCall = httputil.ReqChangeCall
type ReqChangeCalls = httputil.ReqChangeCalls

type ServerSentEvents = xsse.ServerSentEvents
type SSE = xsse.ServerSentEvents

var (
	ContentTypeApplicationJSON = constants.ContentTypeApplicationJSON
	ContentTypeTextPlain       = constants.ContentTypeTextPlain
	ContentTypeUrlencoded      = constants.ContentTypeUrlencoded
)

type Options struct {
	Method         string
	Version        string
	Header         xtypes.SMap
	RespHandler    RespHandler
	SSEHandler     SSEHandler
	SSEOptions     []SSEOption
	SpecifyIP      string
	ReqChangeCalls ReqChangeCalls
}

// WithMethod sets the method for the request.
func WithMethod(method string) RequestOption {
	return func(o *Options) {
		o.Method = method
	}
}

// WithHeaders sets the headers for the request.
func WithHeaders(header map[string]string) RequestOption {
	return func(o *Options) {
		o.Header = header
	}
}

// WithXRequestID sets the request ID for the request.
func WithXRequestID(requestID string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.HeaderRequestId] = requestID
	}
}

// WithContentType sets the content type for the request.
func WithContentType(contentType string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.ContentTypeName] = contentType
	}
}

// WithContentTypeJSON sets the content type to JSON for the request.
func WithContentTypeJSON() RequestOption {
	return WithContentType(ContentTypeApplicationJSON)
}

// WithContentTypeUrlencoded sets the content type to urlencoded for the request.
func WithContentTypeUrlencoded() RequestOption {
	return WithContentType(ContentTypeUrlencoded)
}

// WithRespHandler sets the response handler for the request.
func WithRespHandler(handler RespHandler) RequestOption {
	return func(o *Options) {
		o.RespHandler = handler
	}
}

func WithSSEHandler(handler SSEHandler, opts ...SSEOption) RequestOption {
	return func(o *Options) {
		o.SSEHandler = handler
		o.SSEOptions = opts
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

// WithSpecifyIP 设置指定Ip
func WithSpecifyIP(ip string) RequestOption {
	return func(o *Options) {
		o.SpecifyIP = ip
	}
}

// WithRequestHost 设置修改http.Request对象的自定义方法
func WithReqChangeCall(calls ...ReqChangeCall) RequestOption {
	return func(o *Options) {
		o.ReqChangeCalls = append(o.ReqChangeCalls, calls...)
	}
}
