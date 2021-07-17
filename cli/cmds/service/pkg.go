package service

import (
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/configs"

	"fmt"
	"strings"

	"github.com/zhiyunliu/velocity/compatible"
)

//GetBaseFlags 获取运行时的参数
func GetBaseFlags(cfg *configs.AppSetting) []cli.Flag {
	flags := make([]cli.Flag, 0, 4)
	flags = append(flags, cli.StringFlag{
		Name:        "plat,p",
		Destination: &cfg.PlatName,
		Usage:       "-平台名称",
	})
	flags = append(flags, cli.StringFlag{
		Name:        "system,s",
		Destination: &cfg.SysName,
		Usage:       "-系统名称,默认为当前应用程序名称",
	})
	flags = append(flags, cli.StringFlag{
		Name:        "cluster,c",
		Destination: &cfg.ClusterName,
		Usage:       "-集群名称，默认值为：prod",
	})
	return flags
}

//GetCmdsResult  GetCmdsResult
func GetCmdsResult(serviceName, action string, err error, args ...string) error {
	if err != nil {
		return fmt.Errorf("%s %s %s:%w", action, serviceName, compatible.FAILED, err)
	}
	if len(args) > 0 {
		serviceName = serviceName + " " + strings.Join(args, " ")
	}
	return fmt.Errorf("%s %s %s", action, serviceName, compatible.SUCCESS)
}
