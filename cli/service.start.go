package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/gel/log"
)

//Start Start
func (p *ServiceApp) Start(s service.Service) (err error) {
	log.Infof("serviceApp start:%s", p.cliCtx.App.Name)

	if err = p.loadRegistry(); err != nil {
		return err
	}
	if err = p.loadConfig(); err != nil {
		return err
	}
	log.Info("serviceApp init completed")
	err = p.run()
	if err != nil {
		return err
	}

	log.Infof("serviceApp start:%s completed", p.cliCtx.App.Name)
	return err
}
