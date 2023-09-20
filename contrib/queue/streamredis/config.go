package streamredis

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/golibs/xnet"
)

type ProductOptions struct {
	DelayQueueName       string `json:"delay_queue_name" yaml:"delay_queue_name"`
	StreamMaxLength      int64  `json:"stream_max_len" yaml:"stream_max_len"`           //消息队列长度
	ApproximateMaxLength bool   `json:"approximate_max_len" yaml:"approximate_max_len"` //使用近似队列长度
	RangeSeconds         int    `json:"range_seconds" yaml:"range_seconds"`
	DelayInterval        int    `json:"delay_interval" yaml:"delay_interval"`
}

type ConsumerOptions struct {
	Concurrency       int    `json:"concurrency" yaml:"concurrency"`               //并发的消费者数量
	BufferSize        int    `json:"buffer_size" yaml:"buffer_size"`               //队列长度
	BlockingTimeout   int    `json:"blocking_timeout" yaml:"blocking_timeout"`     //获取消息阻塞时间秒
	VisibilityTimeout int    `json:"visibility_timeout" yaml:"visibility_timeout"` //消息保留时间 秒
	ReclaimInterval   int    `json:"reclaim_interval" yaml:"reclaim_interval"`     //失败消息重试周期
	GroupName         string `json:"group_name" yaml:"group_name"`
}

func getRedisClient(config config.Config) (client *redis.Client, err error) {
	addr := config.Value("addr").String()
	protoType, configName, err := xnet.Parse(addr)
	if err != nil {
		panic(err)
	}
	rootCfg := config.Root()
	queueCfg := rootCfg.Get(protoType).Get(configName)

	client, err = redis.NewByConfig(configName, queueCfg)
	return
}
