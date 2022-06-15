package redis

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/contrib/redis"
	"github.com/zhiyunliu/golibs/xnet"
)

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
