package cron

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/server"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	setting *Setting
	router  *server.RouterGroup
	config  config.Config
	decReq  server.DecodeRequestFunc
	encResp server.EncodeResponseFunc
	encErr  server.EncodeErrorFunc

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() options {
	return options{
		setting: &Setting{
			Config: Config{
				Status: server.StatusStart,
			},
		},
		decReq:  server.DefaultRequestDecoder,
		encResp: server.DefaultResponseEncoder,
		encErr:  server.DefaultErrorEncoder,
		router:  server.NewRouterGroup(""),
	}

}

// WithConfig 设置Config
func WithConfig(config config.Config) Option {
	return func(o *options) {
		o.config = config
	}
}
