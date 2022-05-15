package xhttp

import (
	"github.com/zhiyunliu/gel/constants"
)

type RequestOption func(*Options)

type Options struct {
	Method  string
	Version string
	Header  map[string]string
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
