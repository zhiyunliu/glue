package xhttp

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/golibs/xtypes"
)

type RequestOption func(*Options)

var (
	ContentTypeApplicationJSON = constants.ContentTypeApplicationJSON
	ContentTypeTextPlain       = constants.ContentTypeTextPlain
	ContentTypeUrlencoded      = constants.ContentTypeUrlencoded
)

type Options struct {
	Method  string
	Version string
	Header  xtypes.SMap
}

func WithMethod(method string) RequestOption {
	return func(o *Options) {
		o.Method = method
	}
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

func WithContentType(contentType string) RequestOption {
	return func(o *Options) {
		if o.Header == nil {
			o.Header = make(map[string]string)
		}
		o.Header[constants.ContentTypeName] = contentType
	}
}
