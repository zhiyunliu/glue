package appcli

import (
	"github.com/urfave/cli"
)

//getFlags 获取运行时的参数
func getFlags(cfg *Options) (flags []cli.Flag) {
	flags = append(flags, cli.BoolFlag{
		Name:        "debug,d",
		Destination: &cfg.IsDebug,
		Usage:       `-调试模式，打印更详细的系统运行日志，避免将详细的错误信息返回给调用方`,
	})

	return flags
}
