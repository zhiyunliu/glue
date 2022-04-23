package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/urfave/cli"
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/config/file"
	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/registry"
	"github.com/zhiyunliu/gel/transport"
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xnet"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
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
	cfg := GetSrvConfig(srvApp, args...)
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
func GetSrvConfig(app *ServiceApp, args ...string) *service.Config {
	path, _ := filepath.Abs(os.Args[0])

	svcName := fmt.Sprintf("%s_%s", global.AppName, md5.Str(path)[:8])

	cfg := &service.Config{
		Name:         svcName,
		DisplayName:  global.DisplayName,
		Description:  global.Usage,
		Arguments:    args,
		Dependencies: app.setting.Dependencies,
	}
	cfg.WorkingDirectory = filepath.Dir(path)
	cfg.Option = app.setting.Options
	return cfg
}

//GetSrvApp SrvCfg
func GetSrvApp(c *cli.Context) *ServiceApp {
	opts := c.App.Metadata[options_key].(*Options)

	app := &ServiceApp{
		cliCtx:  c,
		options: opts,
	}
	app.Init()
	return app
}

//ServiceApp ServiceApp
type ServiceApp struct {
	cliCtx     *cli.Context
	options    *Options
	cancelFunc context.CancelFunc
	setting    *appSetting
	instance   *registry.ServiceInstance
	svcCtx     context.Context
}

func (p *ServiceApp) ID() string {
	return p.options.Id
}

func (p *ServiceApp) Metadata() map[string]string {
	return p.options.Metadata
}

func (p *ServiceApp) Endpoint() []string {
	if p.instance == nil {
		return []string{}
	}
	return p.instance.Endpoints
}

func (app *ServiceApp) Init() {
	if app.options.initFile == "" {
		panic(fmt.Errorf("-f 为必须参数"))
	}
	app.options.Config = config.New(config.WithSource(file.NewSource(app.options.initFile)))
	err := app.options.Config.Load()
	if err != nil {
		log.Error("config.Load:%s,Error:%+v", app.options.initFile, err)
	}
	app.loadAppSetting()
	app.loadRegistry()
	app.loadConfig()
}

func (app *ServiceApp) loadAppSetting() {
	setting := &appSetting{}
	err := app.options.Config.Value("app").Scan(setting)
	if err != nil {
		log.Errorf("获取app配置出错:%+v", err)
	}
	app.setting = setting
	global.Mode = app.setting.Mode
	global.LocalIp = xnet.GetLocalIP(setting.IpMask)

}

func (app *ServiceApp) loadRegistry() {

	registrar, err := registry.GetRegistrar(app.options.Config)
	if err != nil {
		log.Error("registry configuration Error:%+v", err)
	}

	app.options.Registrar = registrar
}

func (app *ServiceApp) loadConfig() {
	newSource, err := config.GetConfig(app.options.Config)
	if err != nil {
		log.Errorf("config configuration Error:%+v", err)
	}
	if newSource != nil {
		app.options.Config.Source(newSource)
	}
	global.Config = app.options.Config
}

func (app *ServiceApp) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0)
	for _, e := range app.options.Endpoints {
		endpoints = append(endpoints, e.String())
	}
	if len(endpoints) == 0 {
		for _, srv := range app.options.Servers {
			if r, ok := srv.(transport.Endpointer); ok {
				e := r.Endpoint()
				endpoints = append(endpoints, e.String())
			}
		}
	}
	if app.options.Id == "" {
		app.options.Id = session.Create()
	}
	return &registry.ServiceInstance{
		ID:        app.options.Id,
		Metadata:  app.options.Metadata,
		Name:      global.AppName,
		Version:   global.Version,
		Endpoints: endpoints,
	}, nil
}

type AppInfo interface {
	ID() string
	// Name() string
	// Version() string
	Metadata() map[string]string
	Endpoint() []string
}
