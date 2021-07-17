package server

import "github.com/zhiyunliu/velocity/configs"

//Server 服务器
type Server interface {
	Start(cfg *configs.AppSetting) error
	Shutdown() error
}
