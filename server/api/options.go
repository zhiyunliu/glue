package api

import (
	"net/http"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	serviceName string
	srvCfg      *serverConfig
	logOpts     *log.Options
	config      config.Config
	handler     http.Handler
	router      *engine.RouterGroup
	static      map[string]Static
	decReq      engine.DecodeRequestFunc
	encResp     engine.EncodeResponseFunc
	encErr      engine.EncodeErrorFunc

	startedHooks []engine.Hook
	endHooks     []engine.Hook
}

func setDefaultOption() *options {
	return &options{
		srvCfg: &serverConfig{
			Config: Config{
				Addr:              ":8080",
				Engine:            "gin",
				Status:            engine.StatusStart,
				ReadTimeout:       15,
				WriteTimeout:      15,
				ReadHeaderTimeout: 15,
				MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			},
		},
		logOpts:      &log.Options{},
		static:       make(map[string]Static),
		startedHooks: make([]engine.Hook, 0),
		endHooks:     make([]engine.Hook, 0),
		decReq:       engine.DefaultRequestDecoder,
		encResp:      engine.DefaultResponseEncoder,
		encErr:       engine.DefaultErrorEncoder,
		router:       engine.NewRouterGroup(""),
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
