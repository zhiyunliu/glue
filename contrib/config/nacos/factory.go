package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/contrib/nacos"
	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/log"
)

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return "nacos"
}

func (f *nacosFactory) Create(cfg config.Config) (config.Source, error) {
	ncp, err := nacos.GetClientParam(cfg)
	if err != nil {
		return nil, err
	}
	configClient, err := clients.NewConfigClient(*ncp)
	if err != nil {
		return nil, fmt.Errorf("nacos NewNamingClient error:%+v", err)
	}

	opts := options{
		DataID: global.AppName,
	}

	err = cfg.Value("options").Scan(&opts)
	if err != nil {
		return nil, fmt.Errorf("nacos options error:%+v", err)
	}

	return NewConfigSource(configClient, opts), nil

}

func init() {
	logger.SetLogger(log.DefaultLogger)

	config.Register(&nacosFactory{})
}
