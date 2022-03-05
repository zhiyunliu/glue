package appcli

import "github.com/zhiyunliu/velocity/config"

type Options struct {
	Config  config.Config
	IsDebug bool
	IPMask  string
}

//Option 配置选项
type Option func(*Options)

func WithConfig(cfgVal config.Config) Option {
	return func(o *Options) {
		o.Config = cfgVal
	}
}
