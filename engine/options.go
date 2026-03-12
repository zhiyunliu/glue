package engine

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
)

type Option func(*Options)

type Options struct {
	// SrvType is the server type, such as api, rpc, mqc, cron etc.
	SrvType string
	// SrvName is the server name, 配置文件 servers下的各个服务名称
	SrvName string
	// SvcName is the service name, 注册到注册中心的名称
	SvcName string
	// RequestDecoder is the request decoder function, 解析请求数据
	RequestDecoder DecodeRequestFunc
	// ResponseEncoder is the response encoder function, 编码响应数据
	ResponseEncoder EncodeResponseFunc
	// ErrorEncoder is the error encoder function, 编码错误数据
	ErrorEncoder EncodeErrorFunc
	// LogOpts is the log options, 日志配置
	LogOpts *log.Options
	// Config is the config object, 配置对象
	Config config.Config
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

// WithSrvName is the server name, 配置文件 servers下的各个服务名称
func WithSrvName(name string) Option {
	return func(o *Options) {
		o.SrvName = name
	}
}

// WithSvcName is the service name, 注册到注册中心的名称
func WithSvcName(name string) Option {
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
