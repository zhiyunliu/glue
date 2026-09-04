package sre

import (
	"fmt"

	"github.com/go-kratos/aegis/circuitbreaker/sre"
	"github.com/zhiyunliu/glue/circuitbreaker"
	"github.com/zhiyunliu/glue/config"
)

type provider struct {
	name    string
	opts    *options
	breaker circuitbreaker.CircuitBreaker
}

func (p *provider) Name() string {
	return p.name
}

func (p *provider) CircuitBreaker() circuitbreaker.CircuitBreaker {
	return p.breaker
}
func (p *provider) GetImpl() interface{} {
	return p.breaker
}

type resover struct {
	name string
}

func (r *resover) Name() string {
	return r.name
}

func (r *resover) Resolve(name string, config config.Config) (circuitbreaker.Provider, error) {
	opts := &options{}
	if err := config.ScanTo(opts); err != nil {
		err = fmt.Errorf("sre circuit breaker scan config failed: %w", err)
		return nil, err
	}

	bopts := make([]sre.Option, 0, 4)

	if opts.Success > 0 {
		bopts = append(bopts, sre.WithSuccess(opts.Success))
	}

	if opts.Request > 0 {
		bopts = append(bopts, sre.WithRequest(opts.Request))
	}

	if opts.Bucket > 0 {
		bopts = append(bopts, sre.WithBucket(opts.Bucket))
	}

	if opts.Window > 0 {
		bopts = append(bopts, sre.WithWindow(opts.Window))
	}

	return &provider{
		name:    name,
		opts:    opts,
		breaker: sre.NewBreaker(bopts...),
	}, nil
}

func init() {
	circuitbreaker.Register(&resover{
		name: "sre",
	})
}
