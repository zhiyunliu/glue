package service

import (
	"github.com/kardianos/service"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	if p.server != nil {
		p.server.Shutdown()
	}

	if p.trace != nil {
		p.trace.Stop()
	}
	return nil
}
