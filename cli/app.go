package cli

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/global"
)

const cli_options_key string = "cli_options_key"
const options_key string = "options_key"

//App  cli app
type App struct {
	cliApp     *cli.App
	options    *Options
	cliOptions *cliOptions
}

//Start 启动应用程序
func (a *App) Start() error {
	a.cliApp.Commands = GetCmds(a.cliOptions)
	a.cliApp.Metadata = make(map[string]interface{})
	a.cliApp.Metadata[cli_options_key] = a.cliOptions
	a.cliApp.Metadata[options_key] = a.options
	return a.cliApp.Run(os.Args)
}

//New 创建app
func New(opts ...Option) *App {

	app := &App{options: &Options{}, cliOptions: &cliOptions{}}
	for _, opt := range opts {
		opt(app.options)
	}

	app.cliApp = cli.NewApp()
	app.cliApp.Name = filepath.Base(os.Args[0])
	app.cliApp.Version = global.Version
	app.cliApp.Usage = global.Usage
	cli.HelpFlag = cli.BoolFlag{
		Name:  "help,h",
		Usage: "查看帮助信息",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version,v",
		Usage: "查看版本信息",
	}
	app.cliApp.Metadata = map[string]interface{}{}

	return app
}
