package consul

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"

	"github.com/hashicorp/consul/api"
	"github.com/zhiyunliu/gel/registry"
)

type consulFactory struct {
	opts *options
}

func (f *consulFactory) Name() string {
	return "consul"
}

func (f *consulFactory) Create(cfg config.Config) (registry.Registrar, error) {
	config := api.DefaultConfig()
	opts := &options{
		HealthCheck:         true,
		HealthCheckInterval: 5,
	}

	err := cfg.Scan(opts)
	if err != nil {
		err = fmt.Errorf("consul config error :%+v", err)
		return nil, err
	}

	cliOpts := []Option{}

	cliOpts = append(cliOpts, WithHeartbeat(opts.Heartbeat))
	cliOpts = append(cliOpts, WithHealthCheck(opts.HealthCheck))
	if opts.HealthCheckInterval > 0 {
		cliOpts = append(cliOpts, WithHealthCheckInterval(opts.HealthCheckInterval))
	}

	client, err := api.NewClient(config)
	if err != nil {
		err = fmt.Errorf("consul client error : %+v", err)
		return nil, err
	}
	f.opts = opts
	return New(client, cliOpts...), nil

}

func init() {
	registry.Register(&consulFactory{})
}
