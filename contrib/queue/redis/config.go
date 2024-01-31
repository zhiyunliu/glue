package redis

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xnet"
)

type ProductOptions struct {
	DelayInterval int `json:"delay_interval" yaml:"delay_interval"`
}

func getRedisClient(config config.Config, opts ...queue.Option) (client *redis.Client, err error) {
	addr := config.Value("addr").String()
	protoType, configName, err := xnet.Parse(addr)
	if err != nil {
		panic(err)
	}
	rootCfg := config.Root()
	queueCfg := rootCfg.Get(protoType).Get(configName)

	queueOpts := &queue.Options{}
	for i := range opts {
		opts[i](queueOpts)
	}

	client, err = redis.NewByConfig(configName, queueCfg, queueOpts.CfgData)
	return
}
