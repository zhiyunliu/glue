package velocity

import (
	"github.com/zhiyunliu/velocity/appcli"
	"github.com/zhiyunliu/velocity/compatible"
	"github.com/zhiyunliu/velocity/server"
)

//MicroApp  微服务应用
type MicroApp struct {
	app     *appcli.App
	options []appcli.Option
	manager server.Manager
}

//NewApp 创建微服务应用
func NewApp(manager server.Manager, opts ...appcli.Option) (m *MicroApp) {
	m = &MicroApp{options: opts, manager: manager}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() error {
	m.app = appcli.New(
		m.manager,
		m.options...)

	return m.app.Start()
}

//Close 关闭服务器
func (m *MicroApp) Close() {
	compatible.AppClose()
}
