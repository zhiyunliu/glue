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
	for i := range opts {
		opts[i](redisOpts)
	}
	return newRedis(configName, redisOpts)
}

// NewByConfig 构建客户端
func NewByConfig(configName string, setting config.Config) (r *Client, err error) {
	redisOpts := newOpts()
	setting.Scan(redisOpts)
	return newRedis(configName, redisOpts)
}

func newOpts() *Options {
	return &Options{
		DialTimeout:  5,
		ReadTimeout:  5,
		WriteTimeout: 5,
		PoolSize:     20,
	}
}

func newRedis(configName string, opts *Options) (r *Client, err error) {
	var newOpts = opts
	if Refactor != nil {
		newOpts, err = Refactor(configName, opts)
		if err != nil {
			return
		}
	}

	r = &Client{}
	r.opts = newOpts

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
