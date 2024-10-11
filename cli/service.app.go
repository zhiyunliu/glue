package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kardianos/service"
	"github.com/urfave/cli"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/config/file"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/transport"
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xfile"
	"github.com/zhiyunliu/golibs/xnet"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

var (
	_defaultAppmode AppMode = Release
	_defaultIpMask          = "192.168"
)

type AppService struct {
	service.Service

	srvApp      *ServiceApp
	ServiceName string
	DisplayName string
	Description string
	Arguments   []string
}

// GetService GetServices
func getService(c *cli.Context, args ...string) (srv *AppService, err error) {
	srvApp := GetSrvApp(c)
	if err = srvApp.initApp(); err != nil {
		return
	}
	//1. 构建服务配置
	cfg := GetSrvConfig(srvApp, args...)
	//2.创建本地服务
	appSrv, err := service.New(srvApp, cfg)
	if err != nil {
		return nil, err
	}
	return &AppService{
		Service:     appSrv,
		srvApp:      srvApp,
		ServiceName: cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
	}, err
}

// GetSrvConfig SrvCfg
func GetSrvConfig(app *ServiceApp, args ...string) *service.Config {
	path, _ := filepath.Abs(os.Args[0])

	svcName := fmt.Sprintf("%s_%s", global.AppName, md5.Str(path)[:8])

	cfg := &service.Config{
		Name:         svcName,
		DisplayName:  global.DisplayName,
		Description:  global.Usage,
		Arguments:    args,
		Dependencies: app.options.setting.Dependencies,
	}
	cfg.WorkingDirectory = filepath.Dir(path)
	cfg.Option = app.options.setting.Options
	return cfg
}

// GetSrvApp SrvCfg
func GetSrvApp(c *cli.Context) *ServiceApp {
	opts := c.App.Metadata[options_key].(*Options)
	app := &ServiceApp{
		cliCtx:         c,
		options:        opts,
		closeWaitGroup: &sync.WaitGroup{},
	}
	return app
}

// ServiceApp ServiceApp
type ServiceApp struct {
	cliCtx         *cli.Context
	options        *Options
	instance       *registry.ServiceInstance
	svcCtx         context.Context
	traceEndpoint  *registry.ServerItem
	closeWaitGroup *sync.WaitGroup
}

func (p *ServiceApp) ID() string {
	return p.options.Id
}

func (p *ServiceApp) Metadata() map[string]string {
	return p.options.Metadata
}

func (p *ServiceApp) Endpoint() []registry.ServerItem {
	if p.instance == nil {
		return []registry.ServerItem{}
	}
	return p.instance.Endpoints
}

func (app *ServiceApp) initApp() error {

	if app.options.cmdConfigFile == "" && len(app.options.configSources) == 0 {
		return fmt.Errorf("configFile必须参数")
	}
	if !xfile.Exists(app.options.cmdConfigFile) {
		// global.Mode = string(app.options.setting.Mode)
		// global.LocalIp = xnet.GetLocalIP(app.options.setting.IpMask)
		return fmt.Errorf("config file [%s] 不存在", app.options.cmdConfigFile)
	}
	configSources := app.options.configSources

	absCmdFile, err := filepath.Abs(app.options.cmdConfigFile)
	if err != nil {
		return err
	}
	log.Info("config-file:", absCmdFile)
	configSources = append(configSources, file.NewSource(app.options.cmdConfigFile))

	app.options.Config = config.New(config.WithSource(configSources...))
	err = app.options.Config.Load()
	if err != nil {
		return fmt.Errorf("config.Load:%s,Error:%+v", app.options.cmdConfigFile, err)
	}
	log.Info("serviceApp load appSetting")
	if err = app.loadAppSetting(); err != nil {
		return err
	}
	global.Config = app.options.Config
	return nil
}

func (app *ServiceApp) loadAppSetting() error {
	err := app.options.Config.Value("app").Scan(app.options.setting)
	if err != nil {
		return fmt.Errorf("获取app配置出错:%+v", err)
	}
	global.Mode = string(app.options.setting.Mode)
	global.LocalIp = xnet.GetLocalIP(app.options.setting.IpMask)
	return nil
}

func (app *ServiceApp) loadRegistry() error {
	if app.options.Config == nil {
		return nil
	}
	registrarName := registry.GetRegistrarName(app.options.Config)
	if registrarName == "" {
		return nil
	}
	log.Info("serviceApp load registry:", registrarName)
	registrar, err := registry.GetRegistrar(app.options.Config)
	if err != nil {
		return fmt.Errorf("registry configuration Error:%+v", err)
	}
	app.options.Registrar = registrar
	return nil
}

func (app *ServiceApp) loadConfig() error {
	if app.options.Config == nil {
		return nil
	}
	configName := config.GetConfigName(app.options.Config)
	if configName == "" {
		return nil
	}
	log.Info("serviceApp load config", configName)

	newSource, err := config.GetConfig(app.options.Config)
	if err != nil {
		return fmt.Errorf("get source Error:%+v", err)
	}
	if newSource != nil {
		err = app.options.Config.Source(newSource)
		if err != nil {
			return fmt.Errorf("load Source Error:%+v", err)
		}
	}
	global.Config = app.options.Config
	return nil
}

func (app *ServiceApp) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]registry.ServerItem, 0)
	for _, srv := range app.options.Servers {
		if r, ok := srv.(transport.Endpointer); ok {
			if strings.EqualFold(r.ServiceName(), "") {
				continue
			}
			e := r.Endpoint()
			endpoints = append(endpoints, registry.ServerItem{
				ServiceName: r.ServiceName(),
				EndpointURL: e.String(),
			})
			global.ServerRouterPathList.Store(r.ServiceName(), r.RouterPathList())
		}
	}
	if app.traceEndpoint != nil {
		endpoints = append(endpoints, *app.traceEndpoint)
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
