package bbr

import (
	"fmt"

	aegis "github.com/go-kratos/aegis/ratelimit"
	"github.com/go-kratos/aegis/ratelimit/bbr"
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
	limiter *bbr.BBR
}

func (l *limitwrapper) Allow() (done ratelimit.DoneFunc, err error) {
	innerdone, err := l.limiter.Allow()
	return func(info ratelimit.DoneInfo) {
		innerdone(aegis.DoneInfo{Err: info.Err})
	}, err
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

	bopts := make([]bbr.Option, 0, 4)

	if opts.Window > 0 {
		bopts = append(bopts, bbr.WithWindow(opts.Window))
	}

	if opts.Bucket > 0 {
		bopts = append(bopts, bbr.WithBucket(opts.Bucket))
	}

	if opts.CPUThreshold > 0 {
		bopts = append(bopts, bbr.WithCPUThreshold(int64(opts.CPUThreshold)))
	}

	if opts.CPUQuota > 0 {
		bopts = append(bopts, bbr.WithCPUQuota(opts.CPUQuota))
	}
	return &provider{
		name:    name,
		opts:    opts,
		limiter: &limitwrapper{limiter: bbr.NewLimiter(bopts...)},
	}, nil
}

func init() {
	ratelimit.Register(&resover{})
}
