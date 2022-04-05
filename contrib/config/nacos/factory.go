package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/global"
)

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return "nacos"
}

func (f *nacosFactory) Create(cfg config.Config) (config.Source, error) {

	clientConfig := constant.ClientConfig{}
	serverConfigs := []constant.ServerConfig{}

	err := cfg.Value("client").Scan(&clientConfig)
	if err != nil {
		return nil, fmt.Errorf("nacos client error:%+v", err)
	}
	err = cfg.Value("server").Scan(&serverConfigs)
	if err != nil {
		return nil, fmt.Errorf("nacos server error:%+v", err)
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
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
	config.Register(&nacosFactory{})
}
