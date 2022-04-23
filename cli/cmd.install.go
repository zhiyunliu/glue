package cli

import (
	"os"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/gel/compatible"
)

func init() {
	RegisterFunc(func(cfg *Options) cli.Command {
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
	args := []string{"run", "--nostd"}

	args = append(args, os.Args[2:]...)
	//3.创建本地服务
	srv, err := getService(c, args...)
	if err != nil {
		return err
	}

	err = srv.Install()
	return buildCmdResult(srv.DisplayName, "Install", err)
}
