package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/log"
)

//Start Start
func (p *ServiceApp) Start(s service.Service) (err error) {
	log.Infof("服务启动:%s", p.options.File)
	return p.run()
}
