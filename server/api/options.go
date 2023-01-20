package api

import (
	"net/http"

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
	config      config.Config
	handler     http.Handler
	router      *server.RouterGroup
	static      map[string]Static
	decReq      server.DecodeRequestFunc
	encResp     server.EncodeResponseFunc
	encErr      server.EncodeErrorFunc

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() *options {
	return &options{
		setting: &Setting{
			Config: Config{
				Addr:              ":8080",
				Status:            server.StatusStart,
				ReadTimeout:       15,
				WriteTimeout:      15,
				ReadHeaderTimeout: 15,
				MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			},
		},
		logOpts:      &log.Options{},
		static:       make(map[string]Static),
		startedHooks: make([]server.Hook, 0),
		endHooks:     make([]server.Hook, 0),
		decReq:       server.DefaultRequestDecoder,
		encResp:      server.DefaultResponseEncoder,
		encErr:       server.DefaultErrorEncoder,
		router:       server.NewRouterGroup(""),
	}
}

func WithEndHook(f server.Hook) Option {
	return func(o *options) {
		o.endHooks = append(o.endHooks, f)
	}
}

// WithStartedHook 设置启动回调函数
func WithStartedHook(f server.Hook) Option {
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
