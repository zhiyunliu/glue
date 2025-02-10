package xhttp

import (
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/golibs/httputil"
	"github.com/zhiyunliu/golibs/xtypes"
)

type RequestOption func(*Options)
type RespHandler = httputil.RespHandler

var (
	ContentTypeApplicationJSON = constants.ContentTypeApplicationJSON
	ContentTypeTextPlain       = constants.ContentTypeTextPlain
	ContentTypeUrlencoded      = constants.ContentTypeUrlencoded
)

type Options struct {
	Method  string
	Version string
	Header  xtypes.SMap
	Handler RespHandler
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
		o.Handler = handler
	}
}
