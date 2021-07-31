package appcli

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/appcli/keys"
	"github.com/zhiyunliu/velocity/server"
)

//VERSION 版本号
var VERSION = "1.0.1"

//App  cli app
type App struct {
	*cli.App
	options *Options
}

//Start 启动应用程序
func (a *App) Start() error {
	a.App.Commands = GetCmds(a.options)
	return a.Run(os.Args)
}

//New 创建app
func New(manager server.Manager, opts ...Option) *App {

	app := &App{options: &Options{version: VERSION, usage: "A new cli application"}}
	for _, opt := range opts {
		opt(app.options)
	}

	app.App = cli.NewApp()
	app.App.Name = filepath.Base(os.Args[0])
	app.App.Version = app.options.version
	app.App.Usage = app.options.usage
	cli.HelpFlag = cli.BoolFlag{
		Name:  "help,h",
		Usage: "查看帮助信息",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version,v",
		Usage: "查看版本信息",
	}
	app.App.Metadata = map[string]interface{}{}
	app.App.Metadata[keys.ManagerKey] = manager
	return app
}
