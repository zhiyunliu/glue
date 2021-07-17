package install

import (
	"os"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/compatible"
	"github.com/zhiyunliu/velocity/configs"
)

func init() {
	cmds.RegisterFunc(func(cfg *configs.AppSetting) cli.Command {
		return cli.Command{
			Name:   "install",
			Usage:  "安装服务，以服务方式安装到本地系统",
			Flags:  getFlags(cfg),
			Action: doInstall,
		}
	})
}

func doInstall(c *cli.Context) (err error) {

	//1.检查是否有管理员权限
	if err = compatible.CheckPrivileges(); err != nil {
		return err
	}

	args := []string{"run"}
	args = append(args, os.Args[2:]...)
	//3.创建本地服务
	velocitySrv, err := service.GetService(c, args...)
	if err != nil {
		return err
	}

	err = velocitySrv.Install()
	return service.GetCmdsResult(velocitySrv.DisplayName, "Install", err)
}
