package stop

import (
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/configs"

)

func init() {
	cmds.RegisterFunc(func(cfg *configs.AppSetting) cli.Command {
		return cli.Command{
			Name:   "stop",
			Usage:  "停止服务，停止服务器运行",
			Action: doStop,
		}
	})
}

func doStop(c *cli.Context) (err error) {
	//3.创建本地服务
	velocitySrv, err := service.GetService(c)
	if err != nil {
		return err
	}

	err = velocitySrv.Stop()
	return service.GetCmdsResult(velocitySrv.DisplayName, "Stop", err)
}
