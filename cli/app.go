package cli

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/golibs/xfile"
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
		setting: &appSetting{
			//TraceAddr: _defaultTraceAddr,
			Mode:    _defaultAppmode,
			IpMask:  _defaultIpMask,
			Options: make(map[string]interface{}),
		},
	}}
	for _, opt := range opts {
		opt(app.options)
	}

	fileName := xfile.GetNameWithoutExt(os.Args[0])
	if global.AppName == "" {
		global.AppName = fileName
	}
	if global.DisplayName == "" {
		global.DisplayName = global.AppName
	}

	app.cliApp = cli.NewApp()
	app.cliApp.Name = global.AppName
	app.cliApp.Version = fmt.Sprintf(`
	GitCommit = %s
	BuildTime    = %s
	Version      = %s
	DisplayName  = %s
	GoVersion    = %s
	Usage        = %s
	`,
		global.GitCommit,
		global.BuildTime,
		global.Version,
		global.DisplayName,
		runtime.Version(),
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
	app.cliApp.ExitErrHandler = func(ctx *cli.Context, err error) {
		log.Error(err)
		log.Close()
	}
	return app
}
