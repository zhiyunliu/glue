package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/micro-plat/lib4go/types"
	"github.com/zhiyunliu/velocity/config"
)

//Client redis客户端
type Client struct {
	redis.UniversalClient
	opts *Options
}

//NewByOpts 构建客户端
func NewByOpts(opts ...Option) (r *Client, err error) {
	redisOpts := &Options{}
	for i := range opts {
		opts[i](redisOpts)
	}
	return New(redisOpts)
}

//NewByConfig 构建客户端
func NewByConfig(setting *config.Setting) (r *Client, err error) {
	redisOpts := &Options{}
	setting.Scan(redisOpts)
	return New(redisOpts)
}

func New(opts *Options) (r *Client, err error) {
	r = &Client{}
	r.opts = opts

	r.opts.DialTimeout = types.DecodeInt(r.opts.DialTimeout, 0, 3, r.opts.DialTimeout)
	r.opts.ReadTimeout = types.DecodeInt(r.opts.ReadTimeout, 0, 3, r.opts.ReadTimeout)
	r.opts.WriteTimeout = types.DecodeInt(r.opts.WriteTimeout, 0, 3, r.opts.WriteTimeout)
	r.opts.PoolSize = types.DecodeInt(r.opts.PoolSize, 0, 10, r.opts.PoolSize)

	ropts := &redis.UniversalOptions{
		Addrs:        r.opts.Addrs,
		Password:     r.opts.Password,
		DB:           r.opts.DbIndex,
		DialTimeout:  time.Duration(r.opts.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(r.opts.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(r.opts.WriteTimeout) * time.Second,
		PoolSize:     r.opts.PoolSize,
	}
	r.UniversalClient = redis.NewUniversalClient(ropts)
	_, err = r.UniversalClient.Ping().Result()
	return
}

//GetAddrs GetAddrs
func (c *Client) GetAddrs() []string {
	return c.opts.Addrs
}
