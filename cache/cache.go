package cache

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
	"github.com/zhiyunliu/golibs/xnet"
)

const CacheTypeNode = "caches"

//StandardCache cache
type StandardCache interface {
	GetCache(name string) (q ICache)
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
func (s *xCache) GetCache(name string) (q ICache) {
	obj, err := s.c.GetOrCreate(CacheTypeNode, name, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(CacheTypeNode).Value(name)
		cacheVal := cfgVal.String()
		//redis://localhost
		protoType, configName, err := xnet.Parse(cacheVal)
		if err != nil {
			panic(err)
		}
		cacheCfg := cfg.Get(protoType).Get(configName)
		return newCache(protoType, cacheCfg)
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
	return CacheTypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardCache(c)
}
