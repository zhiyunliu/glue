package redis

import (
	"github.com/zhiyunliu/glue/cache"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/golibs/xnet"
)

func getRedisClient(config config.Config, opts ...cache.Option) (client *redis.Client, err error) {
	addr := config.Value("addr").String()
	protoType, configName, err := xnet.Parse(addr)
	if err != nil {
		panic(err)
	}
	rootCfg := config.Root()
	queueCfg := rootCfg.Get(protoType).Get(configName)

	cacheOpts := &cache.Options{}
	for i := range opts {
		opts[i](cacheOpts)
	}

	client, err = redis.NewByConfig(configName, queueCfg, cacheOpts.CfgData)
	return
}
