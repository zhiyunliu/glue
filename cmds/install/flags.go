package install

import (
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/cmds/service"
	"github.com/zhiyunliu/velocity/configs"
)

var coverIfExists = false

//getFlags 获取运行时的参数
func getFlags() []cli.Flag {
	flags := service.GetBaseFlags()
	flags = append(flags, cli.BoolFlag{
		Name:        "debug,d",
		Destination: &configs.IsDebug,
		Usage:       `-调试模式，打印更详细的系统运行日志，避免将详细的错误信息返回给调用方`,
	})
	flags = append(flags, cli.StringFlag{
		Name:        "trace,t",
		Destination: &configs.TraceType,
		Usage:       `-性能分析。支持:cpu,mem,block,mutex,web`,
	})
	flags = append(flags, cli.StringFlag{
		Name:        "mask,msk",
		Destination: &configs.IPMask,
		Usage:       `-子网掩码。多个网卡情况下根据mask获取本机IP`,
	})
	return flags
}
