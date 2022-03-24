package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/urfave/cli"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/config/file"
	"github.com/zhiyunliu/velocity/global"
	"github.com/zhiyunliu/velocity/log"
	"github.com/zhiyunliu/velocity/registry"
	"github.com/zhiyunliu/velocity/transport"
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
	fileName := filepath.Base(path)
	svcName := fmt.Sprintf("%s_%s", fileName, md5.Str(path)[:8])

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
	CancelFunc context.CancelFunc
	setting    *appSetting
	instance   *registry.ServiceInstance
}

func (p *ServiceApp) ID() string {
	return p.options.Id
}

// func (p *ServiceApp) Name() string {
// 	return p.options.Name
// }

// func (p *ServiceApp) Version() string {
// 	return p.options.Version
// }

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
}

func (app *ServiceApp) loadAppSetting() {
	setting := &appSetting{}
	err := app.options.Config.Value("app").Scan(setting)
	if err != nil {
		log.Errorf("获取app配置出错:%+v", err)
	}
	app.setting = setting
}

func (app *ServiceApp) loadRegistry() {

	registrar, err := registry.GetRegistrar(app.options.Config)
	if err != nil {
		log.Error("registry configuration Error:%+v", err)
	}

	app.options.Registrar = registrar
}

func (app *ServiceApp) loadConfig() {
	newCfg, err := config.GetConfig(app.options.Config)
	if err != nil {
		log.Error("config configuration Error:%+v", err)
	}
	if newCfg != nil {
		app.options.Config = newCfg
	}
}

func (app *ServiceApp) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0, len(app.options.Endpoints))
	for _, e := range app.options.Endpoints {
		endpoints = append(endpoints, e.String())
	}
	if len(endpoints) == 0 {
		for _, srv := range app.options.Servers {
			if r, ok := srv.(transport.Endpointer); ok {
				e, err := r.Endpoint()
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, e.String())
			}
		}
	}
	return &registry.ServiceInstance{
		ID:        app.options.Id,
		Name:      global.DisplayName,
		Version:   global.Version,
		Metadata:  app.options.Metadata,
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
