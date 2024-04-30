package cache

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const (
	TypeNode     = "caches"
	_defaultName = "default"
)

// StandardCache cache
type StandardCache interface {
	GetCache(name string, opts ...Option) (q ICache)
}

// StandardCache cache
type xCache struct {
	c container.Container
}

// NewStandardCache 创建cache
func NewStandardCache(c container.Container) StandardCache {
	return &xCache{c: c}
}

// GetCaches GetCaches
func (s *xCache) GetCache(name string, opts ...Option) (q ICache) {
	realName := name
	if realName == "" {
		realName = _defaultName
	}
	obj, err := s.c.GetOrCreate(TypeNode, realName, func(cfg config.Config) (interface{}, error) {
		//{"proto":"redis","addr":"redis://localhost"}
		cfgVal := cfg.Get(TypeNode).Get(realName)
		protoType := cfgVal.Value("proto").String()
		return newCache(protoType, cfgVal, opts...)
	}, s.getUniqueKey(opts...))
	if err != nil {
		panic(err)
	}
	return obj.(ICache)
}

func (s *xCache) getUniqueKey(opts ...Option) string {
	if len(opts) == 0 {
		return ""
	}
	tmpCfg := &Options{}
	for i := range opts {
		opts[i](tmpCfg)
	}
	return tmpCfg.getUniqueKey()
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
