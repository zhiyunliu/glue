package redis

import (
	"time"

	"github.com/zhiyunliu/gel/cache"
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/contrib/redis"
)

// Redis cache implement
type Redis struct {
	client *redis.Client
}

// connect connect test
func (r *Redis) connect() error {
	var err error
	_, err = r.client.Ping().Result()
	return err
}
func (r *Redis) Name() string {
	return Proto
}

// Get from key
func (r *Redis) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

// Set value with key and expire time
func (r *Redis) Set(key string, val interface{}, expire int) error {
	return r.client.Set(key, val, time.Duration(expire)*time.Second).Err()
}

// Del delete key in redis
func (r *Redis) Del(key string) error {
	return r.client.Del(key).Err()
}

// HashGet from key
func (r *Redis) HashGet(hk, key string) (string, error) {
	return r.client.HGet(hk, key).Result()
}

// HashSet from key
func (r *Redis) HashSet(hk, key string, val string) (bool, error) {
	return r.client.HSet(hk, key, val).Result()
}

// HashDel delete key in specify redis's hashtable
func (r *Redis) HashDel(hk, key string) error {
	return r.client.HDel(hk, key).Err()
}

// Increase
func (r *Redis) Increase(key string) (int64, error) {
	return r.client.Incr(key).Result()
}

func (r *Redis) Decrease(key string) (int64, error) {
	return r.client.Decr(key).Result()
}

// Set ttl
func (r *Redis) Expire(key string, expire int) error {
	return r.client.Expire(key, time.Duration(expire)*time.Second).Err()
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
func (s *redisResolver) Resolve(setting config.Config) (cache.ICache, error) {
	client, err := redis.NewByConfig(setting)
	return &Redis{
		client: client,
	}, err
}
func init() {
	cache.Register(&redisResolver{})
}
