package circuitbreaker

import (
	"fmt"

	_ "github.com/zhiyunliu/glue/circuitbreaker/sre"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const (
	TypeNode = "circuitbreaker"
)

//Standard
type Standard interface {
	GetProvider(name string) (q Provider)
}

//Standard
type xStandrad struct {
	c container.Container
}

//NewStandard
func NewStandard(c container.Container) Standard {
	return &xStandrad{c: c}
}

//GetProvider GetProvider
func (s *xStandrad) GetProvider(name string) (q Provider) {
	if name == "" {
		panic(fmt.Errorf("circuitbreaker provider 配置错误,未设置"))
	}
	obj, err := s.c.GetOrCreate(TypeNode, name, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(name)
		return newProvider(name, cfgVal)
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
	return NewStandard(c)
}
