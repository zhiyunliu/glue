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
	var servers = p.options.Servers
	for i := range p.options.Servers {
		stopErr := servers[i].Stop(p.svcCtx)
		stopMsg := "success"
		if stopErr != nil {
			stopMsg = stopErr.Error()
		}
		log.Infof("stop server[%s]=%s", servers[i].Name(), stopMsg)
		p.closeWaitGroup.Done()
	}
	p.options.StopedHooks.Exec(p.svcCtx, log.DefaultLogger)
	p.closeWaitGroup.Wait()
	log.Infof("serviceApp close:%s stop servers completed", p.cliCtx.App.Name)
}

func (p *ServiceApp) closeLogger() {
	log.Infof("serviceApp close:%s stop logger", p.cliCtx.App.Name)
	log.Close()
}
