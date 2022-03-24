package mqc

import (
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/server"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	setting *Setting
	router  *server.RouterGroup
	config  config.Config
	dec     server.DecodeRequestFunc
	enc     server.EncodeResponseFunc
	ene     server.EncodeErrorFunc

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() options {
	return options{
		dec:    server.DefaultRequestDecoder,
		enc:    server.DefaultResponseEncoder,
		ene:    server.DefaultErrorEncoder,
		router: &server.RouterGroup{},
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
