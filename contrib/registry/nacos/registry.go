package nacos

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/zhiyunliu/glue/registry"
)

var (
	_ registry.Registrar = (*Registry)(nil)
)

type options struct {
	Prefix        string  `json:"prefix"`
	Weight        float64 `json:"weight"`
	Cluster       string  `json:"cluster"`
	Group         string  `json:"group"`
	serverConfigs string  `json:"-"`
}

// Registry is nacos registry.
type Registry struct {
	opts *options
	cli  naming_client.INamingClient
}

// New new a nacos registry.
func New(cli naming_client.INamingClient, opts *options) (r *Registry) {

	return &Registry{
		opts: opts,
		cli:  cli,
	}
}

// Register the registration.
func (r Registry) Name() string {
	return "nacos"
}

func (r Registry) ServerConfigs() string {
	return r.opts.serverConfigs
}

// Register the registration.
func (r Registry) Register(_ context.Context, si *registry.ServiceInstance) error {
	if si.Name == "" {
		return fmt.Errorf("nacos: serviceInstance.name can not be empty")
	}
	for _, item := range si.Endpoints {
		u, err := url.Parse(item.EndpointURL)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		var rmd = map[string]string{}
		if si.Metadata != nil {
			for k, v := range si.Metadata {
				rmd[k] = v
			}
		}
		rmd["scheme"] = u.Scheme
		rmd["version"] = si.Version
		_, e := r.cli.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: item.ServiceName,
			Weight:      r.opts.Weight,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    rmd,
			ClusterName: r.opts.Cluster,
			GroupName:   r.opts.Group,
		})
		if e != nil {
			return fmt.Errorf("RegisterInstance err %v,%v", e, item.EndpointURL)
		}

	}
	return nil
}

// Deregister the registration.
func (r Registry) Deregister(_ context.Context, service *registry.ServiceInstance) error {
	if service == nil {
		return nil
	}
	for _, item := range service.Endpoints {
		u, err := url.Parse(item.EndpointURL)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		if _, err = r.cli.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: item.ServiceName,
			GroupName:   r.opts.Group,
			Cluster:     r.opts.Cluster,
			Ephemeral:   true,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Watch creates a watcher according to the service name.
func (r Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newWatcher(ctx, r.cli, serviceName, r.opts.Group, []string{r.opts.Cluster})
}

// GetService return the service instances in memory according to the service name.
func (r Registry) GetService(_ context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	res, err := r.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   r.opts.Group,
		Clusters:    []string{r.opts.Cluster},
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(res))
	for _, in := range res {
		scheme := in.Metadata["scheme"]
		if scheme == "" {
			scheme = "http"
		}
		items = append(items, &registry.ServiceInstance{
			ID:       in.InstanceId,
			Name:     in.ServiceName,
			Version:  in.Metadata["version"],
			Metadata: in.Metadata,
			Endpoints: []registry.ServerItem{
				{
					ServiceName: serviceName,
					EndpointURL: fmt.Sprintf("%s://%s:%d", scheme, in.Ip, in.Port),
				},
			},
		})
	}
	return items, nil
}

func (r Registry) GetAllServicesInfo(ctx context.Context) (list registry.ServiceList, err error) {
	tmplist, err := r.cli.GetAllServicesInfo(vo.GetAllServiceInfoParam{
		GroupName: r.opts.Group,
		PageNo:    1,
		PageSize:  10000, //默认不超过10000个服务
	})
	if err != nil {
		err = fmt.Errorf("GetAllServicesInfo %w", err)
		return
	}

	list.Count = tmplist.Count
	list.NameList = make([]string, len(tmplist.Doms))
	copy(list.NameList, tmplist.Doms)
	return
}

func (r Registry) GetImpl() any {
	return r.cli
}
