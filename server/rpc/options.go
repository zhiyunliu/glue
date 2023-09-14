package rpc

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/xrpc"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	serviceName  string
	srvCfg       *serverConfig
	logOpts      *log.Options
	router       *engine.RouterGroup
	config       config.Config
	decReq       engine.DecodeRequestFunc
	encResp      engine.EncodeResponseFunc
	encErr       engine.EncodeErrorFunc
	startedHooks []engine.Hook
	endHooks     []engine.Hook
}

func setDefaultOption() *options {
	return &options{
		srvCfg: &serverConfig{
			Config: xrpc.Config{
				Proto:  "grpc",
				Status: engine.StatusStart,
			},
		},
		logOpts: &log.Options{},
		decReq:  engine.DefaultRequestDecoder,
		encResp: engine.DefaultResponseEncoder,
		encErr:  engine.DefaultErrorEncoder,
		router:  engine.NewRouterGroup(""),
	}

}

func WithEndHook(f engine.Hook) Option {
	return func(o *options) {
		o.endHooks = append(o.endHooks, f)
	}
}

// WithStartedHook 设置启动回调函数
func WithStartedHook(f engine.Hook) Option {
	return func(o *options) {
		o.startedHooks = append(o.startedHooks, f)
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
