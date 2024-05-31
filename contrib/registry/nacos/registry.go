package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/nacos_server"
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
	opts        *options
	ncp         *vo.NacosClientParam
	cli         naming_client.INamingClient
	nacosServer *nacos_server.NacosServer
}

// New new a nacos registry.
func New(cli naming_client.INamingClient, ncp *vo.NacosClientParam, opts *options) (r *Registry) {

	return &Registry{
		opts: opts,
		ncp:  ncp,
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
	srvMap := map[string][]string{}

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
		srvMap[item.ServiceName] = append(srvMap[item.ServiceName], item.RouterPathList...)
	}

	for srv, list := range srvMap {
		r.updateServiceMetadata(srv, list)
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

func (r *Registry) getNacosServer() (nacosServer *nacos_server.NacosServer, err error) {
	if r.nacosServer != nil {
		return r.nacosServer, nil
	}
	var (
		clientCfg  constant.ClientConfig   = *r.ncp.ClientConfig
		serverCfgs []constant.ServerConfig = r.ncp.ServerConfigs
	)

	nc := r.cli.(*naming_client.NamingClient).INacosClient

	httpAgent, err := nc.GetHttpAgent()
	if err != nil {
		return nil, err
	}

	nacosServer, err = nacos_server.NewNacosServer(serverCfgs, clientCfg, httpAgent, clientCfg.TimeoutMs, clientCfg.Endpoint)
	if err != nil {
		return nil, err
	}
	r.nacosServer = nacosServer

	return nacosServer, nil
}

func (r *Registry) getSecurityMap(clientConfig *constant.ClientConfig) map[string]string {
	result := make(map[string]string, 2)
	if len(clientConfig.AccessKey) != 0 && len(clientConfig.SecretKey) != 0 {
		result[constant.KEY_ACCESS_KEY] = clientConfig.AccessKey
		result[constant.KEY_SECRET_KEY] = clientConfig.SecretKey
	}
	return result
}

func (r *Registry) updateServiceMetadata(serviceName string, routerList []string) (err error) {
	if len(routerList) == 0 {
		return nil
	}
	nacosServer, err := r.getNacosServer()
	if err != nil {
		return err
	}
	sort.Strings(routerList)
	srv := map[string]string{}
	for i, path := range routerList {
		srv[fmt.Sprintf("router_%d", i+1)] = path
	}

	bytes, _ := json.Marshal(srv)

	//	serviceName=aaa&groupName=DEFAULT_GROUP&namespaceId=
	param := make(map[string]string)
	param["namespaceId"] = r.ncp.ClientConfig.NamespaceId
	param["groupName"] = r.opts.Group
	param["serviceName"] = serviceName
	param["metadata"] = string(bytes)
	param["protectThreshold"] = "0"
	//查看是否存在
	_, err = nacosServer.ReqApi(constant.SERVICE_INFO_PATH, param, http.MethodPut, r.getSecurityMap(r.ncp.ClientConfig))
	return err
}
