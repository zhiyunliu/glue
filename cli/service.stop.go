package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/gel/log"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	err = p.deregister(p.svcCtx)
	if p.cancelFunc != nil {
		log.Infof("serviceApp close:%s stop service-%s", p.cliCtx.App.Name)
		p.cancelFunc()
	}
	log.Infof("serviceApp close:%s stop logger-%s", p.cliCtx.App.Name)
	log.Close()
	return err
}
