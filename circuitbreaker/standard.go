package circuitbreaker

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const (
	TypeNode = "circuitbreaker"
)

var (
	_ Standard = (*xStandrad)(nil)
)

// Standard
type Standard interface {
	GetProvider(proto string) (q Provider)
}

// Standard
type xStandrad struct {
	c container.Container
}

// NewStandard
func NewStandard(c container.Container) Standard {
	return &xStandrad{c: c}
}

// GetProvider GetProvider
func (s *xStandrad) GetProvider(proto string) (q Provider) {
	if proto == "" {
		panic(fmt.Errorf("circuitbreaker provider 配置错误,未设置"))
	}
	obj, err := s.c.GetOrCreate(TypeNode, proto, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(proto)
		return newProvider(proto, cfgVal)
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

func (xBuilder) Build(c container.Container) any {
	return NewStandard(c)
}
