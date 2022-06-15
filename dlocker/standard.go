package dlocker

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
	"github.com/zhiyunliu/golibs/xnet"
)

const TypeNode = "dlocker"

//StandardLocker cache
type StandardLocker interface {
	GetDLocker(name string) (q DLocker)
}

//StandardLocker cache
type xLocker struct {
	c container.Container
}

//NewStandardLocker 创建cache
func NewStandardLocker(c container.Container) StandardLocker {
	return &xLocker{c: c}
}

//GetDLocker GetDLocker
func (s *xLocker) GetDLocker(name string) (q DLocker) {
	obj, err := s.c.GetOrCreate(TypeNode, name, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(TypeNode).Value(name)
		cacheVal := cfgVal.String()
		//redis://localhost
		protoType, configName, err := xnet.Parse(cacheVal)
		if err != nil {
			panic(err)
		}
		cacheCfg := cfg.Get(protoType).Get(configName)
		return newXLocker(protoType, cacheCfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(DLocker)
}

type xBuilder struct{}

func NewBuilder() container.StandardBuilder {
	return &xBuilder{}
}

func (xBuilder) Name() string {
	return TypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardLocker(c)
}
