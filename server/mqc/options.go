package mqc

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/server"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	serviceName string
	setting     *Setting
	logOpts     *log.Options
	router      *server.RouterGroup
	config      config.Config
	decReq      server.DecodeRequestFunc
	encResp     server.EncodeResponseFunc
	encErr      server.EncodeErrorFunc

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() options {
	return options{
		setting: &Setting{
			Config: Config{
				Status: server.StatusStart,
				Addr:   "queues://default",
			},
		},
		logOpts: &log.Options{},
		decReq:  server.DefaultRequestDecoder,
		encResp: server.DefaultResponseEncoder,
		encErr:  server.DefaultErrorEncoder,
		router:  server.NewRouterGroup(""),
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
