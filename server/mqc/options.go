package mqc

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/server"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	serviceName string
	setting     *Setting
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
		setting: &Setting{
			Config: Config{
				Status: server.StatusStart,
				Addr:   "queues://default",
				Engine: "alloter",
			},
			Tasks: TaskList{},
		},
		logOpts: &log.Options{},
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
