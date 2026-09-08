package sre

import (
	"encoding/json"
	"sync"

	"github.com/go-kratos/aegis/circuitbreaker/sre"
	"github.com/zhiyunliu/glue/circuitbreaker"
	"github.com/zhiyunliu/glue/config"
)

var (
	_ circuitbreaker.Provider = (*provider)(nil)
)

type provider struct {
	name        string
	breakerMap  sync.Map
	mutexlocker sync.Mutex
}

func (p *provider) Name() string {
	return p.name
}

func (p *provider) Build(cachekey string, opts ...circuitbreaker.Option) (ciruitbreaker circuitbreaker.CircuitBreaker) {
	if breaker, ok := p.breakerMap.Load(cachekey); ok {
		return breaker.(circuitbreaker.CircuitBreaker)
	}
	p.mutexlocker.Lock()
	defer p.mutexlocker.Unlock()

	if breaker, ok := p.breakerMap.Load(cachekey); ok {
		return breaker.(circuitbreaker.CircuitBreaker)
	}

	copts := &circuitbreaker.Options{}
	for _, opt := range opts {
		opt(copts)
	}

	options := &options{}
	if len(copts.ConfigData) != 0 {
		_ = json.Unmarshal(copts.ConfigData, options)
	}

	sreOpts := make([]sre.Option, 0, 4)
	if options.Success > 0 {
		sreOpts = append(sreOpts, sre.WithSuccess(options.Success))
	}

	if options.Request > 0 {
		sreOpts = append(sreOpts, sre.WithRequest(options.Request))
	}

	if options.Bucket > 0 {
		sreOpts = append(sreOpts, sre.WithBucket(options.Bucket))
	}

	if options.Window > 0 {
		sreOpts = append(sreOpts, sre.WithWindow(options.Window))
	}
	ciruitbreaker = sre.NewBreaker(sreOpts...)
	return ciruitbreaker
}

type resover struct {
	name string
}

func (r *resover) Name() string {
	return r.name
}

func (r *resover) Resolve(name string, config config.Config) (circuitbreaker.Provider, error) {
	return &provider{
		name:        name,
		breakerMap:  sync.Map{},
		mutexlocker: sync.Mutex{},
	}, nil
}

func init() {
	circuitbreaker.Register(&resover{
		name: "sre",
	})
}
