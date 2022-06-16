package cli

import (
	"github.com/urfave/cli"
)

//getFlags 获取运行时的参数
func getFlags(cfg *Options) (flags []cli.Flag) {
	flags = append(flags,
		cli.StringFlag{
			Name:        "file,f",
			Destination: &cfg.initFile,
			Usage:       `-配置文件`,
			Value:       "config.json",
		},
	)

	return flags
}
