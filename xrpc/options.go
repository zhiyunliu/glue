package xrpc

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/xtypes"
)

// StreamType is the type of stream.
type StreamType int

const (
	UnknownStream       StreamType = 0
	BidirectionalStream StreamType = 1
	ClientStream        StreamType = 2
	ServerStream        StreamType = 3
)

type RequestOption func(*Options)

type Options struct {
	Header             xtypes.SMap // 请求头
	Method             string      // 请求方法
	Query              string      // 请求参数
	WaitForReady       bool        // 是否等待服务端响应
	MaxCallRecvMsgSize int         // 最大接收消息体大小，默认4M(maximum message size in bytes the client can receive)
	MaxCallSendMsgSize int         // 最大发送消息体大小，默认4M(maximum message size in bytes the client can send)
	StreamProcessor    any         // 是否使用流传输
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

// WithStreamProcessor 设置使用流传输
func WithStreamProcessor(processor any) RequestOption {
	return func(o *Options) {
		o.StreamProcessor = processor
	}
}

// WithStreamDefaultProcessor 设置使用默认流传输
func WithStreamDefaultProcessor() RequestOption {
	return func(o *Options) {
		o.StreamProcessor = DefaultProcessor{}
	}
}

// MaxCallRecvMsgSize
func MaxCallRecvMsgSize(size int) RequestOption {
	return func(o *Options) {
		o.MaxCallRecvMsgSize = size
	}
}

// MaxCallSendMsgSize
func MaxCallSendMsgSize(size int) RequestOption {
	return func(o *Options) {
		o.MaxCallSendMsgSize = size
	}
}

// StreamRecvOptions is a struct that contains options for receiving messages.
type StreamRecvOptions struct {
	Unmarshal StreamUnmarshaler
}

// StreamRevcOption is a function that sets an option for receiving messages.
type StreamRevcOption func(*StreamRecvOptions)

// WithStreamUnmarshal sets the callback function to unmarshal a received message.
func WithStreamUnmarshal(callback StreamUnmarshaler) StreamRevcOption {
	return func(sro *StreamRecvOptions) {
		sro.Unmarshal = callback
	}
}
