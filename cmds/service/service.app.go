package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/configs"
	"github.com/zhiyunliu/velocity/libs"
	"github.com/zhiyunliu/velocity/server"
)

type AppService struct {
	service.Service
	ServiceName string
	DisplayName string
	Description string
	Arguments   []string
}

//GetService GetService
func GetService(c *cli.Context, args ...string) (velocitySrv *AppService, err error) {
	//1. 构建服务配置
	cfg := GetSrvConfig(args...)

	//2.创建本地服务
	appSrv, err := service.New(GetSrvApp(c), cfg)
	if err != nil {
		return nil, err
	}
	return &AppService{
		Service:     appSrv,
		ServiceName: cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
	}, err
}

//GetSrvConfig SrvCfg
func GetSrvConfig(args ...string) *service.Config {
	path, _ := filepath.Abs(os.Args[0])

	svcName := fmt.Sprintf("%s_%s", configs.SysName, libs.Md5(path)[:8])

	cfg := &service.Config{
		Name:         svcName,
		DisplayName:  svcName,
		Description:  configs.Usage,
		Arguments:    args,
		Dependencies: []string{"After=network.target syslog.target"},
	}
	cfg.WorkingDirectory = filepath.Dir(path)
	// cfg.Option = make(map[string]interface{})
	// cfg.Option["LimitNOFILE"] = 10240
	return cfg
}

//GetSrvApp SrvCfg
func GetSrvApp(c *cli.Context) *ServiceApp {
	return &ServiceApp{
		c: c,
	}
}

//ServiceApp ServiceApp
type ServiceApp struct {
	c      *cli.Context
	server server.ResponsiveServer
	trace  itrace
}
