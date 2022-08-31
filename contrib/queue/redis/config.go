package redis

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/golibs/xnet"
)

type ProductOptions struct {
	DelayQueueName string `json:"delay_queue_name" yaml:"delay_queue_name"`
	RangeSeconds   int    `json:"range_seconds" yaml:"range_seconds"`
	DelayInterval  int    `json:"delay_interval" yaml:"delay_interval"`
}

func getRedisClient(config config.Config) (client *redis.Client, err error) {
	addr := config.Value("addr").String()
	protoType, configName, err := xnet.Parse(addr)
	if err != nil {
		panic(err)
	}
	rootCfg := config.Root()
	queueCfg := rootCfg.Get(protoType).Get(configName)

	client, err = redis.NewByConfig(queueCfg)
	return
}
