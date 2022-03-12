package appcli

import (
	"github.com/urfave/cli"
)

func init() {
	RegisterFunc(func(cfg *cliOptions) cli.Command {
		return cli.Command{
			Name:   "remove",
			Usage:  "删除服务，从本地服务器移除服务",
			Action: doRemove,
		}
	})
}
func doRemove(c *cli.Context) (err error) {

	//3.创建本地服务
	srv, err := getService(c)
	if err != nil {
		return err
	}
	err = srv.Uninstall()
	return buildCmdResult(srv.DisplayName, "Remove", err)

}
