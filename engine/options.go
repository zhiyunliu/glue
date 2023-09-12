package engine

import (
	"github.com/zhiyunliu/glue/log"
)

type Option func(*Options)

type Options struct {
	SrvType         string
	SrvName         string
	RequestDecoder  DecodeRequestFunc  //:          server.DefaultRequestDecoder,
	ResponseEncoder EncodeResponseFunc //:          server.DefaultResponseEncoder,
	ErrorEncoder    EncodeErrorFunc    //:          server.DefaultErrorEncoder,
	LogOpts         *log.Options
}

func DefaultOptions() *Options {
	return &Options{
		RequestDecoder:  DefaultRequestDecoder,
		ResponseEncoder: DefaultResponseEncoder,
		ErrorEncoder:    DefaultErrorEncoder,
	}
}

func WithSrvType(srvType string) Option {
	return func(o *Options) {
		o.SrvType = srvType
	}
}

func WithSrvName(name string) Option {
	return func(o *Options) {
		o.SrvName = name
	}
}
func WithRequestDecoder(requestDecoder DecodeRequestFunc) Option {
	return func(o *Options) {
		o.RequestDecoder = requestDecoder
	}
}

func WithResponseEncoder(responseEncoder EncodeResponseFunc) Option {
	return func(o *Options) {
		o.ResponseEncoder = responseEncoder
	}
}

func WithErrorEncoder(errorEncoder EncodeErrorFunc) Option {
	return func(o *Options) {
		o.ErrorEncoder = errorEncoder
	}
}

func WithLogOptions(opt *log.Options) Option {
	return func(o *Options) {
		o.LogOpts = opt
	}
}
