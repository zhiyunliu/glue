package mqc

import "github.com/zhiyunliu/velocity/server"

// Option 参数设置类型
type Option func(*options)

type options struct {
	setting *Setting

	router *server.RouterGroup

	startedHooks []server.Hook
	endHooks     []server.Hook
}

func setDefaultOption() options {
	return options{
		router: &server.RouterGroup{},
	}

}
