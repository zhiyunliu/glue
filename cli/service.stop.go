package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/gel/log"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	if p.CancelFunc != nil {
		p.CancelFunc()
	}

	err = p.deregister(p.svcCtx)
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	log.Infof("服务关闭:%s %s", p.cliCtx.App.Name, msg)

	log.Close()
	return nil
}
