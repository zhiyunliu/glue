package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/gel/log"
)

//Start Start
func (p *ServiceApp) Start(s service.Service) (err error) {
	log.Infof("服务启动:%s", p.cliCtx.App.Name)

	log.Info("serviceApp load registry")
	if err = p.loadRegistry(); err != nil {
		return err
	}
	log.Info("serviceApp load config")
	if err = p.loadConfig(); err != nil {
		return err
	}
	log.Info("serviceApp init completed")

	err = p.run()
	if err != nil {
		return err
	}

	log.Infof("服务启动:%s completed", p.cliCtx.App.Name)
	return err
}
