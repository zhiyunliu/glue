package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/nacos"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
)

const _name = "nacos"

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return _name
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
		DataID: fmt.Sprintf("%s.json", global.AppName),
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
