package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/log"
)

//Start Start
func (p *ServiceApp) Start(s service.Service) (err error) {
	log.Infof("服务启动:%s", p.cliCtx.App.Name)
	err = p.run()
	log.Infof("服务启动:%s completed", p.cliCtx.App.Name)
	return err
}
