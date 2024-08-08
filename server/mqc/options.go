package mqc

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/xmqc"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	serviceName string
	srvCfg      *serverConfig
	logOpts     *log.Options
	router      *engine.RouterGroup
	config      config.Config
	decReq      engine.DecodeRequestFunc
	encResp     engine.EncodeResponseFunc
	encErr      engine.EncodeErrorFunc

	startedHooks []engine.Hook
	endHooks     []engine.Hook
}

func setDefaultOption() *options {
	return &options{
		srvCfg: &serverConfig{
			Config: xmqc.Config{
				Status: engine.StatusStart,
				Addr:   "queues://default",
				Proto:  "alloter",
			},
		},
		logOpts: &log.Options{
			WithSource:  &[]bool{true}[0],
			WithHeaders: []string{constants.HeaderAuthorization, constants.HeaderReferer, constants.HeaderXForwardedFor},
		},
		decReq:  engine.DefaultRequestDecoder,
		encResp: engine.DefaultResponseEncoder,
		encErr:  engine.DefaultErrorEncoder,
		router:  engine.NewRouterGroup(""),
	}

}

// WithStartedHook 设置启动回调函数
func WithConfig(config config.Config) Option {
	return func(o *options) {
		o.config = config
	}
}

// WithServiceName 设置服务名称
func WithServiceName(serviceName string) Option {
	return func(o *options) {
		o.serviceName = serviceName
	}
}

// Log 设置日志配置
func Log(opts ...log.ServerOption) Option {
	return func(o *options) {
		for i := range opts {
			opts[i](o.logOpts)
		}
	}
}

// WithDecodeRequestFunc 解析入参
func WithDecodeRequestFunc(decReq engine.DecodeRequestFunc) Option {
	return func(o *options) {
		o.decReq = decReq
	}
}

// WithEncodeResponseFunc 编码响应
func WithEncodeResponseFunc(encResp engine.EncodeResponseFunc) Option {
	return func(o *options) {
		o.encResp = encResp
	}
}

// WithEncodeErrorFunc 编码错误
func WithEncodeErrorFunc(encErr engine.EncodeErrorFunc) Option {
	return func(o *options) {
		o.encErr = encErr
	}
}
