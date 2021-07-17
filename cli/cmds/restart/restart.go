package restart

import (
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/configs"
)

var isFixed bool

func init() {
	cmds.RegisterFunc(func(cfg *configs.AppSetting) cli.Command {
		return cli.Command{
			Name:   "restart",
			Usage:  "重启服务",
			Action: doRestart,
		}
	})
}

func doRestart(c *cli.Context) (err error) {

	//3.创建本地服务
	velocitySrv, err := service.GetService(c)
	if err != nil {
		return err
	}
	err = velocitySrv.Restart()
	return err
}
