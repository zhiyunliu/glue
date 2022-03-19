package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/log"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	if p.CancelFunc != nil {
		p.CancelFunc()
	}
	log.Debugf("服务关闭:%s", p.cliCtx.App.Name)
	return nil
}
