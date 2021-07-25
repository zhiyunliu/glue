package service

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/logger"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	if p.server != nil {
		p.server.Shutdown()
	}

	if p.trace != nil {
		p.trace.Stop()
	}

	logger.Infof("服务关闭:%s", p.config.Addr)
	return nil
}
