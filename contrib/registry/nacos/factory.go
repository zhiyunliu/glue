package nacos

import (
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/contrib/nacos"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/registry"
)

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return "nacos"
}

func (f *nacosFactory) Create(cfg config.Config) (registry.Registrar, error) {

	ncp, err := nacos.GetClientParam(cfg)
	if err != nil {
		return nil, err
	}

	namingClient, err := clients.NewNamingClient(*ncp)
	if err != nil {
		return nil, fmt.Errorf("nacos NewNamingClient error:%+v", err)
	}

	opts := &options{
		Prefix:  "/microservices",
		Cluster: "DEFAULT",
		Group:   constant.DEFAULT_GROUP,
		Weight:  100,
	}

	err = cfg.Value("options").Scan(opts)
	if err != nil {
		return nil, fmt.Errorf("nacos options error:%+v", err)
	}

	addrs := make([]string, 0)

	for _, s := range ncp.ServerConfigs {
		addrs = append(addrs, fmt.Sprintf("%s:%d", s.IpAddr, s.Port))
	}

	opts.serverConfigs = strings.Join(addrs, ",")

	return New(namingClient, opts), nil

}

func init() {
	logger.SetLogger(log.DefaultLogger)

	registry.Register(&nacosFactory{})
}
