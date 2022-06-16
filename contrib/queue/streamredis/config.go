package streamredis

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/golibs/xnet"
)

type ProductOptions struct {
	StreamMaxLength int64 `json:"stream_max_len" yaml:"stream_max_len"`
}

type ConsumerOptions struct {
	Concurrency     int `json:"concurrency" yaml:"concurrency"`
	BufferSize      int `json:"buffer_size" yaml:"buffer_size"`
	BlockingTimeout int `json:"blocking_timeout" yaml:"blocking_timeout"`
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
