package run

import (
	"os"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/globals"
)

func init() {
	cmds.RegisterFunc(func(cfg *globals.AppSetting) cli.Command {
		flags := getFlags(cfg)
		return cli.Command{
			Name:   "run",
			Usage:  "运行服务,以前台方式运行服务。通过终端输出日志，终端关闭后服务自动退出。",
			Flags:  flags,
			Action: doRun,
		}
	})
}

//doRun 服务启动
func doRun(c *cli.Context) (err error) {
	//1.创建本地服务
	velocitySrv, err := service.GetService(c, os.Args[2:]...)
	if err != nil {
		return err
	}
	err = velocitySrv.Run()
	return err
}
