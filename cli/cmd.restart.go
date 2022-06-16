package cli

import (
	"github.com/urfave/cli"
)

func init() {
	RegisterFunc(func(cfg *Options) cli.Command {
		return cli.Command{
			Name:   "restart",
			Usage:  "重启服务",
			Flags:  getFlags(cfg),
			Action: doRestart,
		}
	})
}

func doRestart(c *cli.Context) (err error) {

	//3.创建本地服务
	srv, err := getService(c)
	if err != nil {
		return err
	}
	err = srv.Restart()
	return err
}
