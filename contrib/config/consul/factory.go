package consul

import (
	"log"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
)

const _name = "consul"

type consulFactory struct {
}

func (f *consulFactory) Name() string {
	return _name
}

func (f *consulFactory) Create(cfg config.Config) (config.Source, error) {
	cltCfg := &clientConfig{}
	cfg.Scan(cltCfg)

	config := consulapi.DefaultConfig()
	config.Address = cltCfg.Addr
	config.Datacenter = cltCfg.Addr
	config.Namespace = cltCfg.Addr
	config.Partition = cltCfg.Addr
	config.Scheme = cltCfg.Addr
	config.Token = cltCfg.Addr
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	opts := []Option{}
	opts = append(opts, WithPath(global.AppName))

	return New(client, opts...)

}

func init() {

	config.Register(&consulFactory{})
}
