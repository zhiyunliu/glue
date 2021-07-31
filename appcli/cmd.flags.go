package appcli

import (
	"github.com/urfave/cli"
)

//getFlags 获取运行时的参数
func getFlags(cfg *Options) (flags []cli.Flag) {
	flags = append(flags, cli.StringFlag{
		Name:        "plat,p",
		Destination: &cfg.PlatName,
		Usage:       "-平台名称",
	})
	flags = append(flags, cli.BoolFlag{
		Name:        "debug,d",
		Destination: &cfg.IsDebug,
		Usage:       `-调试模式，打印更详细的系统运行日志，避免将详细的错误信息返回给调用方`,
	})
	flags = append(flags, cli.StringFlag{
		Name:        "trace,t",
		Destination: &cfg.TraceType,
		Usage:       `-性能分析。支持:cpu,mem,block,mutex,web`,
	})
	flags = append(flags, cli.StringFlag{
		Name:        "tport,tp",
		Destination: &cfg.TracePort,
		Usage:       `-性能分析服务端口号。用于trace为web模式时的端口号。默认：19999`,
	})

	return flags
}
