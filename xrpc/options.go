package xrpc

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/golibs/xtypes"
)

type RequestOption func(*Options)

type Options struct {
	Header       xtypes.SMap // map[string]string
	WaitForReady bool
}

func WithHeaders(header map[string]string) RequestOption {
	return func(o *Options) {
		o.Header = header
	}
}

func WithXRequestID(requestID string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.HeaderRequestId] = requestID
	}
}
func WithWaitForReady(waitForReady bool) RequestOption {
	return func(o *Options) {
		o.WaitForReady = waitForReady
	}
}

func WithContentType(contentType string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.ContentTypeName] = contentType
	}
}
