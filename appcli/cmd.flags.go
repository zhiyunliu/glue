package appcli

import (
	"github.com/urfave/cli"
)

//getFlags 获取运行时的参数
func getFlags(cfg *cliOptions) (flags []cli.Flag) {
	flags = append(flags,
		cli.BoolFlag{
			Name:        "debug,d",
			Destination: &cfg.IsDebug,
			Usage:       `-调试模式，打印更详细的系统运行日志，避免将详细的错误信息返回给调用方`,
		},
		cli.StringFlag{
			Name:        "file,f",
			Destination: &cfg.File,
			Usage:       `-配置文件`,
		},
		cli.StringFlag{
			Name:        "registry,r",
			Destination: &cfg.Registry,
			Usage:       `-注册中心地址。格式：proto://host。如：zk://ip1,ip2`,
		},
		cli.StringFlag{
			Name:        "mask,mask",
			Destination: &cfg.IPMask,
			Usage:       `-子网掩码。多个网卡情况下根据mask筛选本机IP`,
		})

	return flags
}
