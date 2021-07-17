package service

import (
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/configs"
)

//GetBaseFlags 获取运行时的参数
func GetBaseFlags() []cli.Flag {
	flags := make([]cli.Flag, 0, 4)
	flags = append(flags, platFlag)
	flags = append(flags, sysNameFlag)
	flags = append(flags, clusterFlag)
	return flags
}

var platFlag = cli.StringFlag{
	Name:        "plat,p",
	Destination: &configs.PlatName,
	Usage:       "-平台名称",
}

var sysNameFlag = cli.StringFlag{
	Name:        "system,s",
	Destination: &configs.SysName,
	Usage:       "-系统名称,默认为当前应用程序名称",
}

var clusterFlag = cli.StringFlag{
	Name:        "cluster,c",
	Destination: &configs.ClusterName,
	Usage:       "-集群名称，默认值为：prod",
}
