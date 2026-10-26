package redis

import (
	"context"
	"time"

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
	const minTimeoutCheckInterval = time.Millisecond * 100
	opt := &dlocker.Options{
		Reentrant:            true,
		TimeoutCheckInterval: minTimeoutCheckInterval,
	}
	for i := range opts {
		opts[i](opt)
	}
	if opt.TimeoutCheckInterval < minTimeoutCheckInterval {
		opt.TimeoutCheckInterval = minTimeoutCheckInterval
	}
	return newLock(r, key, opt)
}

// Eval 执行脚本
func (r *Redis) Eval(ctx context.Context, cmd string, keys []string, vals []string) (obj interface{}, err error) {

	args := make([]interface{}, len(vals))
	for i := range vals {
		args[i] = vals[i]
	}

	return r.client.Eval(ctx, cmd, keys, args...).Result()
}

type redisResolver struct {
}

func (s *redisResolver) Name() string {
	return Proto
}
func (s *redisResolver) Resolve(configName string, setting config.Config) (dlocker.DLockerBuilder, error) {
	// 加锁脚本可重入但非幂等，禁止网络错误重试，并让命令遵循获取锁的截止时间。
	client, err := redis.NewByConfig(configName, setting, map[string]any{
		"context_timeout_enabled": true,
		"max_retries":             -1,
	})
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
