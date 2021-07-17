package velocity

import (
	"github.com/lib4dev/cli"
	"github.com/zhiyunliu/velocity/compatible"
	"github.com/zhiyunliu/velocity/configs"
	"github.com/zhiyunliu/velocity/server"
)

//MicroApp  微服务应用
type MicroApp struct {
	app    *cli.App
	config *configs.AppSetting
}

//NewApp 创建微服务应用
func NewApp(server server.ResponsiveServer, opts ...Option) (m *MicroApp) {
	m = &MicroApp{config: &configs.AppSetting{}}
	for _, opt := range opts {
		opt(m.config)
	}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() {
	m.app = cli.New(cli.WithVersion(configs.Version), cli.WithUsage(configs.Usage))
	m.app.Start()
}

//Close 关闭服务器
func (m *MicroApp) Close() {
	Close()
}

//Close 关闭服务器
func Close() {
	compatible.AppClose()
}
