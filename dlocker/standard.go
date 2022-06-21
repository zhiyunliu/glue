package dlocker

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
	"github.com/zhiyunliu/golibs/xnet"
)

const TypeNode = "dlocker"

//StandardLocker cache
type StandardLocker interface {
	GetDLocker() (q DLockerBuilder)
}

//StandardLocker lock
type xLocker struct {
	c container.Container
}

//NewStandardLocker 创建 lock
func NewStandardLocker(c container.Container) StandardLocker {
	return &xLocker{c: c}
}

//GetDLocker GetDLocker
func (s *xLocker) GetDLocker() (q DLockerBuilder) {
	obj, err := s.c.GetOrCreate(TypeNode, TypeNode, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Value(TypeNode).String()
		//redis://localhost
		protoType, configName, err := xnet.Parse(cfgVal)
		if err != nil {
			return nil, err
		}
		cacheCfg := cfg.Get(protoType).Get(configName)
		return newXLocker(protoType, cacheCfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(DLockerBuilder)
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
