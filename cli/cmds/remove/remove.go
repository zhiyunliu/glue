package remove

import (
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/configs"

	"github.com/urfave/cli"
)

func init() {
	cmds.RegisterFunc(func(cfg *configs.AppSetting) cli.Command {
		return cli.Command{
			Name:   "remove",
			Usage:  "删除服务，从本地服务器移除服务",
			Action: doRemove,
		}
	})
}
func doRemove(c *cli.Context) (err error) {

	//3.创建本地服务
	velocitySrv, err := service.GetService(c)
	if err != nil {
		return err
	}
	err = velocitySrv.Uninstall()
	return service.GetCmdsResult(velocitySrv.DisplayName, "Remove", err)

}
