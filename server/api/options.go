package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/server"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	setting *Setting
	handler http.Handler
	router  *server.RouterGroup
	dec     server.DecodeRequestFunc
	enc     server.EncodeResponseFunc
	ene     server.EncodeErrorFunc

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() *options {
	return &options{
		handler:      gin.New(),
		startedHooks: make([]server.Hook, 0),
		endHooks:     make([]server.Hook, 0),
		dec:          server.DefaultRequestDecoder,
		enc:          server.DefaultResponseEncoder,
		ene:          server.DefaultErrorEncoder,

		router: server.NewRouterGroup(),
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
		setting := &Setting{}
		config.Scan(setting)
		o.setting = setting
	}
}
