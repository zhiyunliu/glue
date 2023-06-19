package metrics

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const (
	TypeNode = "metrics"
)

// StandardMetric
type StandardMetric interface {
	GetProvider(name string) (q Provider)
}

// StandardMetric
type xMetric struct {
	c container.Container
}

// NewStandardMetric
func NewStandardMetric(c container.Container) StandardMetric {
	return &xMetric{c: c}
}

// GetProvider GetProvider
func (s *xMetric) GetProvider(protoName string) (q Provider) {
	if protoName == "" {
		panic(fmt.Errorf("metric provider 配置错误,未设置"))
	}
	obj, err := s.c.GetOrCreate(protoName, protoName, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(protoName)
		return newProvider(protoName, cfgVal)
	})
	if err != nil {
		panic(err)
	}
	return obj.(Provider)
}

type xBuilder struct{}

func NewBuilder() container.StandardBuilder {
	return &xBuilder{}
}

func (xBuilder) Name() string {
	return TypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardMetric(c)
}
