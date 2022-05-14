package cli

import (
	"github.com/kardianos/service"
	"github.com/zhiyunliu/gel/log"
)

//Stop Stop
func (p *ServiceApp) Stop(s service.Service) (err error) {
	log.Infof("服务关闭:%s %s-%s", p.cliCtx.App.Name, "关闭注册中心", "开始")
	err = p.deregister(p.svcCtx)
	msg := "完成"
	if err != nil {
		msg = err.Error()
	}
	log.Infof("服务关闭:%s %s-%s", p.cliCtx.App.Name, "关闭注册中心", msg)

	log.Infof("服务关闭:%s %s-%s", p.cliCtx.App.Name, "关闭服务", "开始")
	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	log.Infof("服务关闭:%s %s-%s", p.cliCtx.App.Name, "关闭服务", "完成")
	log.Close()
	return nil
}
