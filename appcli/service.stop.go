package appcli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/log"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	if p.CancelFunc != nil {
		p.CancelFunc()
	}
	log.Infof("服务关闭:%s", p.manager.Name())
	return nil
}
