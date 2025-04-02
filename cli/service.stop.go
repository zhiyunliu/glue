package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/glue/log"
)

// Stop Stop
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

	p.options.StopingHooks.Exec(p.svcCtx, log.DefaultLogger)

	for i := range p.options.Servers {
		if err := p.options.Servers[i].Stop(p.svcCtx); err != nil {
			log.Errorf("stop server %s error:%s", p.options.Servers[i].Name(), err.Error())
		}
		p.closeWaitGroup.Done()
	}
	p.options.StopedHooks.Exec(p.svcCtx, log.DefaultLogger)
	p.closeWaitGroup.Wait()
}

func (p *ServiceApp) closeLogger() {
	log.Infof("serviceApp close:%s stop logger", p.cliCtx.App.Name)
	log.Close()
}
