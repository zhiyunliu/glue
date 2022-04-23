package cli

import (
	"github.com/urfave/cli"
)

func init() {
	RegisterFunc(func(cfg *Options) cli.Command {

		return cli.Command{
			Name:   "start",
			Usage:  "启动服务，以后台方式运行服务",
			Flags:  getFlags(cfg),
			Action: doStart,
		}
	})
}

func doStart(c *cli.Context) (err error) {
	//3.创建本地服务
	srv, err := getService(c)
	if err != nil {
		return err
	}
	err = srv.Start()
	return buildCmdResult(srv.DisplayName, "Start", err)
}
