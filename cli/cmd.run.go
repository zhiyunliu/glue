package cli

import (
	"os"

	"github.com/urfave/cli"
)

func init() {
	RegisterFunc(func(cfg *Options) cli.Command {
		flags := getFlags(cfg)
		return cli.Command{
			Name:  "run",
			Usage: "运行服务,以前台方式运行服务。通过终端输出日志，终端关闭后服务自动退出。",
			Flags: append(flags, cli.BoolFlag{
				Name:  "nostd",
				Usage: `关闭std输出`,
			}),
			Action: doRun,
		}
	})
}

//doRun 服务启动
func doRun(c *cli.Context) (err error) {

	srv, err := getService(c, os.Args[2:]...)
	if err != nil {
		return err
	}
	return srv.Run()
}
