package engine

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
)

type Option func(*Options)

type Options struct {
	SrvType         string
	SrvName         string
	RequestDecoder  DecodeRequestFunc
	ResponseEncoder EncodeResponseFunc
	ErrorEncoder    EncodeErrorFunc
	LogOpts         *log.Options
	Config          config.Config
}

func DefaultOptions() *Options {
	return &Options{
		RequestDecoder:  DefaultRequestDecoder,
		ResponseEncoder: DefaultResponseEncoder,
		ErrorEncoder:    DefaultErrorEncoder,
		Config:          global.Config,
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

func WithConfig(cfg config.Config) Option {
	return func(o *Options) {
		o.Config = cfg
	}
}
