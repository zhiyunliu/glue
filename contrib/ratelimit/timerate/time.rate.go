package timerate

import (
	"fmt"

	"golang.org/x/time/rate"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/ratelimit"
)

type provider struct {
	name    string
	opts    *options
	limiter *limitwrapper
}

func (p *provider) Name() string {
	return "bbr"
}
func (p *provider) GetImpl() interface{} {
	return p.limiter.limiter
}

func (p *provider) Limiter() ratelimit.Limiter {
	return p.limiter
}

type limitwrapper struct {
	limiter *rate.Limiter
}

func (l *limitwrapper) Allow() (done ratelimit.DoneFunc, err error) {
	isAllow := l.limiter.Allow()
	if !isAllow {
		return nil, ratelimit.ErrNotAllow
	}
	return func(info ratelimit.DoneInfo) {}, nil
}

type resover struct {
}

func (r *resover) Name() string {
	return "bbr"
}

func (r *resover) Resolve(name string, config config.Config) (ratelimit.Provider, error) {
	opts := &options{}
	if err := config.ScanTo(opts); err != nil {
		err = fmt.Errorf("ratelimit bbr scan config failed: %w", err)
		return nil, err
	}

	limiter := rate.NewLimiter(rate.Limit(float64(opts.CountLimit)/float64(opts.Every)), opts.CountLimit)

	return &provider{
		name:    name,
		opts:    opts,
		limiter: &limitwrapper{limiter: limiter},
	}, nil
}

func init() {
	ratelimit.Register(&resover{})
}
