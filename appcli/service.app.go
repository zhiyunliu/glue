package appcli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/appcli/keys"
	"github.com/zhiyunliu/velocity/libs/security"
	"github.com/zhiyunliu/velocity/server"
)

type AppService struct {
	service.Service
	ServiceName string
	DisplayName string
	Description string
	Arguments   []string
}

//GetService GetServices
func getService(c *cli.Context, args ...string) (srv *AppService, err error) {
	srvApp := GetSrvApp(c)
	//1. 构建服务配置
	cfg := GetSrvConfig(srvApp.options, args...)
	//2.创建本地服务
	appSrv, err := service.New(srvApp, cfg)
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
func GetSrvConfig(opts *Options, args ...string) *service.Config {
	path, _ := filepath.Abs(os.Args[0])

	svcName := fmt.Sprintf("%s_%s", opts.SysName, security.Md5(path)[:8])

	cfg := &service.Config{
		Name:         svcName,
		DisplayName:  svcName,
		Description:  opts.Usage,
		Arguments:    args,
		Dependencies: []string{"After=network.target syslog.target"},
	}
	cfg.WorkingDirectory = filepath.Dir(path)
	cfg.Option = make(map[string]interface{})
	cfg.Option["LimitNOFILE"] = 10240
	return cfg
}

//GetSrvApp SrvCfg
func GetSrvApp(c *cli.Context) *ServiceApp {
	server := c.App.Metadata[keys.ManagerKey].(server.Manager)
	opts := c.App.Metadata[keys.OptionsKey].(*Options)
	initAppConfig(opts)
	return &ServiceApp{
		cliCtx:  c,
		manager: server,
		options: opts,
	}
}

//ServiceApp ServiceApp
type ServiceApp struct {
	cliCtx     *cli.Context
	manager    server.Manager
	options    *Options
	CancelFunc context.CancelFunc
}

func initAppConfig(config *Options) {
	if config.Addr == "" {
		config.Addr = ":8081"
	}

	if config.PlatName == "" {
		config.PlatName = "default"
	}
	if config.SysName == "" {
		config.SysName = filepath.Base(os.Args[0])
	}

	if config.Version == "" {
		config.Version = "1.0.0"
	}
}
