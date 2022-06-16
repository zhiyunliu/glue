package cache

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const (
	TypeNode     = "caches"
	_defaultName = "default"
)

//StandardCache cache
type StandardCache interface {
	GetCache(name ...string) (q ICache)
}

//StandardCache cache
type xCache struct {
	c container.Container
}

//NewStandardCache 创建cache
func NewStandardCache(c container.Container) StandardCache {
	return &xCache{c: c}
}

//GetCaches GetCaches
func (s *xCache) GetCache(name ...string) (q ICache) {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}

	obj, err := s.c.GetOrCreate(TypeNode, realName, func(cfg config.Config) (interface{}, error) {
		//{"proto":"redis","addr":"redis://localhost"}
		cfgVal := cfg.Get(TypeNode).Get(realName)
		protoType := cfgVal.Value("proto").String()
		return newCache(protoType, cfgVal)

	})
	if err != nil {
		panic(err)
	}
	return obj.(ICache)
}

type xBuilder struct{}

func NewBuilder() container.StandardBuilder {
	return &xBuilder{}
}

func (xBuilder) Name() string {
	return TypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardCache(c)
}
