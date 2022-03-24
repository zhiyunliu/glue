package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/registry"
)

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return "nacos"
}

func (f *nacosFactory) Create(cfg config.Config) (registry.Registrar, error) {

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

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("nacos NewNamingClient error:%+v", err)
	}

	clientOpts := options{}

	err = cfg.Value("options").Scan(&clientOpts)
	if err != nil {
		return nil, fmt.Errorf("nacos options error:%+v", err)
	}

	return New(namingClient, clientOpts.toOpts()...), nil

}

func init() {
	registry.Register(&nacosFactory{})
}
