package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/gel/global"
)

const options_key string = "options_key"

//App  cli app
type App struct {
	cliApp  *cli.App
	options *Options
}

//Start 启动应用程序
func (a *App) Start() error {
	a.cliApp.Commands = GetCmds(a.options)
	a.cliApp.Metadata = make(map[string]interface{})
	a.cliApp.Metadata[options_key] = a.options
	return a.cliApp.Run(os.Args)
}

//New 创建app
func New(opts ...Option) *App {

	app := &App{options: &Options{
		RegistrarTimeout: 10 * time.Second,
		StopTimeout:      10 * time.Second,
	}}
	for _, opt := range opts {
		opt(app.options)
	}

	app.cliApp = cli.NewApp()
	app.cliApp.Name = filepath.Base(os.Args[0])
	app.cliApp.Version = fmt.Sprintf(`
	GitCommitLog = %s
	BuildTime    = %s
	Version      = %s
	GoVersion    = %s
	DisplayName  = %s
	Usage        = %s
	`,
		global.GitCommitLog,
		global.BuildTime,
		global.Version,
		global.GoVersion,
		global.DisplayName,
		global.Usage,
	)
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
