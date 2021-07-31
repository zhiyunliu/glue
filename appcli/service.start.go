package appcli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/logger"
)

//Start Start
func (p *ServiceApp) Start(s service.Service) (err error) {
	logger.Infof("服务启动:%s", p.manager.Name())
	return p.run()
}
