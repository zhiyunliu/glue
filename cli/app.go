package cli

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/server"
	"github.com/zhiyunliu/velocity/configs"
)

//VERSION 版本号
var VERSION = "0.0.1"

//App  cli app
type App struct {
	*cli.App
	option *option
}

//Start 启动应用程序
func (a *App) Start() error {
	return a.Run(os.Args)
}

//New 创建app
func New(server server.Server, config *configs.AppSetting, opts ...Option) *App {

	app := &App{option: &option{version: VERSION, usage: "A new cli application"}}
	for _, opt := range opts {
		opt(app.option)
	}

	app.App = cli.NewApp()
	app.App.Name = filepath.Base(os.Args[0])
	app.App.Version = app.option.version
	app.App.Usage = app.option.usage
	cli.HelpFlag = cli.BoolFlag{
		Name:  "help,h",
		Usage: "查看帮助信息",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version,v",
		Usage: "查看版本信息",
	}
	app.App.Metadata = map[string]interface{}{}
	app.App.Metadata["server"] = server
	app.App.Metadata["config"] = config
	app.App.Commands = cmds.GetCmds(config)
	return app
}
