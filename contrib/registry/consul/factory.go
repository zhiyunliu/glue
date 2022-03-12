package consul

import (
	"fmt"
	"strconv"

	"github.com/zhiyunliu/velocity/log"

	"github.com/hashicorp/consul/api"
	"github.com/zhiyunliu/velocity/registry"
)

type consulFactory struct {
}

func (f *consulFactory) Name() string {
	return "consul"
}

func (f *consulFactory) Create(opts *registry.Options) (registry.Registrar, error) {
	config := api.DefaultConfig()

	if addr, ok := opts.Metadata.Get("addr"); ok {
		config.Address = fmt.Sprint(addr)
	}
	cliOpts := []Option{}

	if tbv, ok := opts.Metadata.Get("enableHeartbeat"); ok {
		bv, _ := strconv.ParseBool(fmt.Sprint(tbv))
		cliOpts = append(cliOpts, WithHeartbeat(bv))
	}
	if tbv, ok := opts.Metadata.Get("enableHealthCheck"); ok {
		bv, _ := strconv.ParseBool(fmt.Sprint(tbv))
		cliOpts = append(cliOpts, WithHealthCheck(bv))
	}
	if tiv, ok := opts.Metadata.Get("healthcheckInterval"); ok {
		iv, _ := strconv.ParseInt(fmt.Sprint(tiv), 10, 32)
		cliOpts = append(cliOpts, WithHealthCheckInterval(int(iv)))
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}
	return New(client, cliOpts...), nil

}

func init() {
	registry.Register(&consulFactory{})
}
