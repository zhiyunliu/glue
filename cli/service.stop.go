package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/gel/log"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	err = p.deregister(p.svcCtx)
	if err != nil {
		return err
	}
	p.stopServers()
	p.closeLogger()
	return nil
}

func (p *ServiceApp) stopServers() {
	log.Infof("serviceApp close:%s stop servers", p.cliCtx.App.Name)
	for i := range p.options.Servers {
		p.options.Servers[i].Stop(p.svcCtx)
		p.closeWaitGroup.Done()
	}
	p.closeWaitGroup.Wait()
}

func (p *ServiceApp) closeLogger() {
	log.Infof("serviceApp close:%s stop logger", p.cliCtx.App.Name)
	log.Close()
}
