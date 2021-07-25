package velocity

import (
	"os"
	"path/filepath"

	"github.com/zhiyunliu/velocity/cli"
	"github.com/zhiyunliu/velocity/server"
	"github.com/zhiyunliu/velocity/compatible"
 	"github.com/zhiyunliu/velocity/globals"

	_ "github.com/zhiyunliu/velocity/cli/cmds/install"
	_ "github.com/zhiyunliu/velocity/cli/cmds/remove"
	_ "github.com/zhiyunliu/velocity/cli/cmds/restart"
	_ "github.com/zhiyunliu/velocity/cli/cmds/run"
	_ "github.com/zhiyunliu/velocity/cli/cmds/start"
	_ "github.com/zhiyunliu/velocity/cli/cmds/status"
	_ "github.com/zhiyunliu/velocity/cli/cmds/stop"
)

//MicroApp  微服务应用
type MicroApp struct {
	app    *cli.App
	config *globals.AppSetting
	server server.Server
}

//NewApp 创建微服务应用
func NewApp(server server.Server, opts ...Option) (m *MicroApp) {
	m = &MicroApp{config: &globals.AppSetting{
		SysName: filepath.Base(os.Args[0]),
	}, server: server}
	for _, opt := range opts {
		opt(m.config)
	}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() error {
	m.app = cli.New(
		m.server,
		m.config,
		cli.WithVersion(m.config.Version),
		cli.WithUsage(m.config.Usage))
	return m.app.Start()
}

//Close 关闭服务器
func (m *MicroApp) Close() {
	compatible.AppClose()
}
