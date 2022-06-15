package rpc

import (
	"math"

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
				Addr:           ":28080",
				Status:         server.StatusStart,
				MaxRecvMsgSize: math.MaxInt32,
				MaxSendMsgSize: math.MaxInt32,
			},
		},
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
