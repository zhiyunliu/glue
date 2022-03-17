package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/zhiyunliu/velocity/extlib/xtypes"
	"github.com/zhiyunliu/velocity/registry"
)

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return "nacos"
}

func (f *nacosFactory) Create(opts *registry.Options) (registry.Registrar, error) {
	clientConfig := &constant.ClientConfig{}
	if clientCfg, ok := opts.Metadata.Get("client"); ok {
		txmap := clientCfg.(xtypes.XMap)
		txmap.Scan(clientCfg)
	}

	serverConfigs := []constant.ServerConfig{}
	if serverCfg, ok := opts.Metadata.Get("server"); ok {
		txmap := serverCfg.(xtypes.XMap)
		txmap.Scan(&serverConfigs)
	}
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, err
	}

	clientOpts := options{}
	if optCfg, ok := opts.Metadata.Get("options"); ok {
		txmap := optCfg.(xtypes.XMap)
		txmap.Scan(&opts)
	}

	return New(namingClient, clientOpts.toOpts()...), nil

}

func init() {
	registry.Register(&nacosFactory{})
}
