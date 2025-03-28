package prometheus

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/metrics"
)

const (
	Proto = "prometheus"
)

var (
	_ metrics.Provider = &xProvider{}
)

type xProvider struct {
	factory metrics.Factory
	config  config.Config
}

func (p xProvider) Name() string {
	return Proto
}

func (p *xProvider) Build(metric any) error {
	return metrics.Init(metric, p.factory, p.config)
}

func (p *xProvider) GetImpl() any {
	return p.factory
}
