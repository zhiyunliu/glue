package run

import (
	"os"

	"github.com/lib4dev/cli/cmds"
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cmds/service"
)

func init() {
	cmds.RegisterFunc(func() cli.Command {
		flags := getFlags()
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
