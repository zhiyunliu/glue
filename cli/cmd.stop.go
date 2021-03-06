package cli

import (
	"github.com/urfave/cli"
)

func init() {
	RegisterFunc(func(cfg *Options) cli.Command {
		return cli.Command{
			Name:   "stop",
			Usage:  "停止服务，停止服务器运行",
			Flags:  getFlags(cfg),
			Action: doStop,
		}
	})
}

func doStop(c *cli.Context) (err error) {
	//3.创建本地服务
	srv, err := getService(c)
	if err != nil {
		return err
	}

	err = srv.Stop()
	return buildCmdResult(srv.DisplayName, "Stop", err)
}
