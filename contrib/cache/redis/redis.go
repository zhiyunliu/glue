package redis

import (
	"context"
	"time"

	rds "github.com/go-redis/redis/v7"
	"github.com/zhiyunliu/glue/cache"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/redis"
)

// Redis cache implement
type Redis struct {
	client *redis.Client
}

func (r *Redis) Name() string {
	return Proto
}

// Get from key
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	v, err := r.client.Get(key).Result()
	if err == rds.Nil {
		return "", cache.Nil
	}
	return v, err
}

// Set value with key and expire time
func (r *Redis) Set(ctx context.Context, key string, val interface{}, expire int) error {
	err := r.client.Set(key, val, time.Duration(expire)*time.Second).Err()
	if err == rds.Nil {
		return cache.Nil
	}
	return err
}

// Del delete key in redis
func (r *Redis) Del(ctx context.Context, key string) error {
	err := r.client.Del(key).Err()
	if err == rds.Nil {
		return cache.Nil
	}
	return err
}

// HashGet from key
func (r *Redis) HashGet(ctx context.Context, hk, key string) (string, error) {
	v, err := r.client.HGet(hk, key).Result()
	if err == rds.Nil {
		return "", cache.Nil
	}
	return v, err
}

// HashSet from key
func (r *Redis) HashSet(ctx context.Context, hk, key string, val string) (bool, error) {
	v, err := r.client.HSet(hk, key, val).Result()
	if err == rds.Nil {
		return v > 0, cache.Nil
	}
	return v > 0, err
}

// HashDel delete key in specify redis's hashtable
func (r *Redis) HashDel(ctx context.Context, hk, key string) error {
	err := r.client.HDel(hk, key).Err()
	if err == rds.Nil {
		return cache.Nil
	}
	return err
}

func (r *Redis) HashMGet(ctx context.Context, hk string, key ...string) (map[string]interface{}, error) {
	vals, err := r.client.HMGet(hk, key...).Result()
	result := make(map[string]interface{}, len(vals))
	if len(vals) > 0 {
		for i := range key {
			result[key[i]] = vals[i]
		}
	}
	return result, err
}
func (r *Redis) HashSetAll(ctx context.Context, hk string, val map[string]interface{}) (bool, error) {
	return r.client.HMSet(hk, val).Result()
}

func (r *Redis) HashExists(ctx context.Context, hk, key string) (bool, error) {
	return r.client.HExists(hk, key).Result()
}

// Increase
func (r *Redis) Increase(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(key).Result()
}

func (r *Redis) Decrease(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(key).Result()
}

// Set ttl
func (r *Redis) Expire(ctx context.Context, key string, expire int) error {
	err := r.client.Expire(key, time.Duration(expire)*time.Second).Err()
	if err == rds.Nil {
		return cache.Nil
	}
	return err
}

// Exists
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	v, err := r.client.Exists(key).Result()
	if err == rds.Nil {
		return v > 0, cache.Nil
	}
	return v > 0, err
}

// GetImpl 暴露原生client
func (r *Redis) GetImpl() interface{} {
	return r.client
}

type redisResolver struct {
}

func (s *redisResolver) Name() string {
	return Proto
}
func (s *redisResolver) Resolve(config config.Config, opts ...cache.Option) (cache.ICache, error) {
	client, err := getRedisClient(config, opts...)
	if err != nil {
		return nil, err
	}
	return &Redis{
		client: client,
	}, err
}
func init() {
	cache.Register(&redisResolver{})
}
