package redis

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/dlocker"
)

// Redis cache implement
type Redis struct {
	client *redis.Client
}

// GetImpl 暴露原生client
func (r *Redis) GetImpl() interface{} {
	return r.client
}

// Build 构建锁
func (r *Redis) Build(key string, opts ...dlocker.Option) dlocker.DLocker {
	opt := &dlocker.Options{}
	for i := range opts {
		opts[i](opt)
	}

	return newLock(r, key, opt)
}

// Eval 执行脚本
func (r *Redis) Eval(cmd string, keys []string, vals []string) (obj interface{}, err error) {

	args := make([]interface{}, len(vals))
	for i := range vals {
		args[i] = vals[i]
	}

	return r.client.Eval(cmd, keys, args...).Result()
}

type redisResolver struct {
}

func (s *redisResolver) Name() string {
	return Proto
}
func (s *redisResolver) Resolve(configName string, setting config.Config) (dlocker.DLockerBuilder, error) {
	client, err := redis.NewByConfig(configName, setting, nil)
	if err != nil {
		return nil, err
	}
	return &Redis{
		client: client,
	}, err

}
func init() {
	dlocker.Register(&redisResolver{})
}
