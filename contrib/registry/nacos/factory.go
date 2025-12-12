package nacos

import (
	"fmt"
	"os"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/nacos"
	"github.com/zhiyunliu/glue/registry"
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

	err = cfg.Value("options").ScanTo(opts)
	if err != nil {
		return nil, fmt.Errorf("nacos options error:%+v", err)
	}
	if len(opts.Group) == 0 {
		opts.Group = os.Getenv("NACOS_DISCOVERY_GROUP")
	}
	if len(opts.Cluster) == 0 {
		opts.Cluster = os.Getenv("NACOS_DISCOVERY_CLUSTER")
	}
	addrs := make([]string, 0)

	for _, s := range ncp.ServerConfigs {
		addrs = append(addrs, fmt.Sprintf("%s:%d", s.IpAddr, s.Port))
	}

	opts.serverConfigs = strings.Join(addrs, ",")

	if len(opts.Group) == 0 {
		opts.Group = constant.DEFAULT_GROUP
	}

	return New(namingClient, ncp, opts), nil

}

func init() {
	registry.Register(&nacosFactory{})
}

func buildServiceInstanceList(serviceName string, instances []model.Instance) []*registry.ServiceInstance {

	items := make([]*registry.ServiceInstance, 0, len(instances))
	for _, in := range instances {
		scheme := in.Metadata["scheme"]
		if scheme == "" {
			scheme = "http"
		}
		rmd := make(map[string]string, len(in.Metadata)+2)
		rmd["scheme"] = scheme
		rmd["cluster"] = in.ClusterName
		for k, v := range in.Metadata {
			rmd[k] = v
		}

		items = append(items, &registry.ServiceInstance{
			ID:       in.InstanceId,
			Name:     in.ServiceName,
			Version:  in.Metadata["version"],
			Metadata: rmd,
			Endpoints: []registry.ServerItem{
				{
					ServiceName: serviceName,
					EndpointURL: fmt.Sprintf("%s://%s:%d", scheme, in.Ip, in.Port),
				},
			},
		})
	}
	return items
}
