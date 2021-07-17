package service

import "github.com/kardianos/service"

//Start Start
func (p *ServiceApp) Start(s service.Service) (err error) {
	return p.run()
}
