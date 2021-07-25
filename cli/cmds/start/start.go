package start

import (
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/globals"

	"github.com/urfave/cli"
)

 
func init() {
	cmds.RegisterFunc(func(cfg *globals.AppSetting) cli.Command {

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
