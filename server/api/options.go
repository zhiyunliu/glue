package api

import (
	"net/http"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/server"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	setting *Setting
	config  config.Config
	handler http.Handler
	router  *server.RouterGroup
	static  map[string]Static
	decReq  server.DecodeRequestFunc
	encResp server.EncodeResponseFunc
	encErr  server.EncodeErrorFunc

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() *options {
	return &options{
		setting:      &Setting{},
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
