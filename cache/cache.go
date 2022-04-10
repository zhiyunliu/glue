package cache

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
	"github.com/zhiyunliu/golibs/xnet"
)

const cacheTypeNode = "caches"

//StandardCache cache
type StandardCache struct {
	c container.Container
}

//NewStandardCache 创建cache
func NewStandardCache(c container.Container) *StandardCache {
	return &StandardCache{c: c}
}

//GetCaches GetCaches
func (s *StandardCache) GetCache(name string) (q ICache) {
	obj, err := s.c.GetOrCreate(cacheTypeNode, name, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(cacheTypeNode).Value(name)
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
