package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/zhiyunliu/velocity/global"
	"github.com/zhiyunliu/velocity/libs/types"
	"github.com/zhiyunliu/velocity/registry"
	"github.com/zhiyunliu/velocity/server"
)

type nacosClient struct {
	opts         *registry.Options
	namingClient naming_client.INamingClient
}

func (c *nacosClient) Register(server server.Runnable) error {
	metadata := types.XMap{}
	metadata.Merge(c.opts.Metadata)
	metadata.Merge(server.Metadata())
	_, err := c.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          global.LocalIp,
		Port:        server.Port(),
		ServiceName: server.Name(),
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata.SMap(),    //map[string]string{"idc": "shanghai"},
		ClusterName: global.ClusterName, // "cluster-a",       // default value is DEFAULT
		GroupName:   global.GroupName,   // default value is DEFAULT_GROUP
	})
	return err
}
func (c *nacosClient) Deregister() error {
	_, err := c.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          "10.0.0.11",
		Port:        8848,
		ServiceName: "demo.go",
		Ephemeral:   true,
		Cluster:     "cluster-a", // default value is DEFAULT
		GroupName:   "group-a",   // default value is DEFAULT_GROUP
	})
	return err
}

func (c *nacosClient) GetImpl() interface{} {
	return c.namingClient
}

func createNacosClient(opts *registry.Options) (registry.IRegistry, error) {

	regClient := &nacosClient{
		opts: opts,
	}

	clientConfig := &constant.ClientConfig{}

	if clientCfg, ok := opts.Metadata.Get("client"); ok {
		txmap := clientCfg.(types.XMap)
		txmap.Scan(clientCfg)
	}

	serverConfigs := []constant.ServerConfig{}
	if serverCfg, ok := opts.Metadata.Get("server"); ok {
		txmap := serverCfg.(types.XMap)
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
	regClient.namingClient = namingClient
	return regClient, nil
}
