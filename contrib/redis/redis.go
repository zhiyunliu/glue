package redis

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/golibs/xtransform"
)

var Nil = redis.Nil

// Client redis客户端
type Client struct {
	configName string
	redis.UniversalClient
	opts *Options
}

// NewByOpts 构建客户端
func NewByOpts(configName string, opts ...Option) (r *Client, err error) {
	redisOpts := defaultRedisOpts()
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
	redisOpts := defaultRedisOpts()
	setting.ScanTo(redisOpts)
	if Refactor != nil {
		redisOpts, err = Refactor(configName, redisOpts)
		if err != nil {
			return
		}
	}
	return newRedis(configName, redisOpts, mapCfg)
}

func defaultRedisOpts() *Options {
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

	r = &Client{
		configName: configName,
	}
	opts.Username = xtransform.TranslateCallback(opts.Username, func(param string) string {
		val := os.Getenv(param)
		if len(val) > 0 {
			return val
		}
		return param
	}, xtransform.WithBraceMode(), xtransform.WithAtBraceMode())

	opts.Password = xtransform.TranslateCallback(opts.Password, func(param string) string {
		val := os.Getenv(param)
		if len(val) > 0 {
			return val
		}
		return param
	}, xtransform.WithBraceMode(), xtransform.WithAtBraceMode())
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
	_, err = r.UniversalClient.Ping(context.Background()).Result()
	return
}

// GetAddrs GetAddrs
func (c *Client) GetAddrs() []string {
	return c.opts.Addrs
}
