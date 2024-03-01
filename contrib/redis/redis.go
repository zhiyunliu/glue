package redis

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/zhiyunliu/glue/config"
)

var Nil = redis.Nil

// Client redis客户端
type Client struct {
	redis.UniversalClient
	opts *Options
}

// NewByOpts 构建客户端
func NewByOpts(configName string, opts ...Option) (r *Client, err error) {
	redisOpts := newOpts()
	if Refactor != nil {
		redisOpts, err = Refactor(configName, redisOpts)
		if err != nil {
			return
		}
	}
	for i := range opts {
		opts[i](redisOpts)
	}
	return newRedis(configName, redisOpts, map[string]any{})
}

// NewByConfig 构建客户端
func NewByConfig(configName string, setting config.Config, mapCfg map[string]any) (r *Client, err error) {
	redisOpts := newOpts()
	setting.ScanTo(redisOpts)
	if Refactor != nil {
		redisOpts, err = Refactor(configName, redisOpts)
		if err != nil {
			return
		}
	}
	return newRedis(configName, redisOpts, mapCfg)
}

func newOpts() *Options {
	return &Options{
		DialTimeout:  5,
		ReadTimeout:  5,
		WriteTimeout: 5,
		PoolSize:     20,
	}
}

func newRedis(configName string, opts *Options, mapCfg map[string]any) (r *Client, err error) {
	if len(mapCfg) > 0 {
		WithMapConfig(mapCfg)(opts)
	}

	r = &Client{}
	r.opts = opts

	ropts := &redis.UniversalOptions{
		Addrs:        r.opts.Addrs,
		Username:     r.opts.Username,
		Password:     r.opts.Password,
		DB:           int(r.opts.DbIndex),
		DialTimeout:  time.Duration(r.opts.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(r.opts.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(r.opts.WriteTimeout) * time.Second,
		PoolSize:     int(r.opts.PoolSize),
	}
	r.UniversalClient = redis.NewUniversalClient(ropts)
	_, err = r.UniversalClient.Ping().Result()
	return
}

// GetAddrs GetAddrs
func (c *Client) GetAddrs() []string {
	return c.opts.Addrs
}
