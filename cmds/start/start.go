package start

import (
	"github.com/lib4dev/cli/cmds"
	"github.com/zhiyunliu/velocity/cmds/service"

	"github.com/urfave/cli"
)

var isFixed bool

func init() {
	cmds.RegisterFunc(func() cli.Command {

		return cli.Command{
			Name:   "start",
			Usage:  "启动服务，以后台方式运行服务",
			Action: doStart,
		}
	})
}

func doStart(c *cli.Context) (err error) {
	//3.创建本地服务
	velocitySrv, err := service.GetService(c)
	if err != nil {
		return err
	}
	err = velocitySrv.Start()
	return service.GetCmdsResult(velocitySrv.DisplayName, "Start", err)
}
