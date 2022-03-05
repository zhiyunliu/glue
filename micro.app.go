package velocity

import (
	"fmt"

	"github.com/zhiyunliu/velocity/appcli"
	"github.com/zhiyunliu/velocity/compatible"
	"github.com/zhiyunliu/velocity/server"
)

//MicroApp  微服务应用
type MicroApp struct {
	cliApp     *appcli.App
	manager    server.Manager
	serverList map[string]server.Runnable
}

//NewApp 创建微服务应用
func NewApp() (m *MicroApp) {
	m = &MicroApp{serverList: make(map[string]server.Runnable)}
	return m
}

func (m *MicroApp) AddServer(server server.Runnable) {
	m.serverList[server.Name()] = server
}

//Start 启动服务器
func (m *MicroApp) Start() error {
	if len(m.serverList) == 0 {
		return fmt.Errorf("没有需要启动都服务应用")
	}
	m.cliApp = appcli.New(m.manager)

	return m.cliApp.Start()
}

//Close 关闭服务器
func (m *MicroApp) Close() {
	compatible.AppClose()
}
