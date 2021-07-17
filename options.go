package velocity

import "github.com/zhiyunliu/velocity/configs"

type Option func(cfg *configs.AppSetting)

func WithPlatName(platName string) Option {
	return func(cfg *configs.AppSetting) {
		configs.PlatName = platName
	}
}
