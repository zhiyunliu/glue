package cache

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
)

const cacheTypeNode = "caches"

//StandardCaches cache
type StandardCaches struct {
	c container.IContainer
}

//NewStandardCaches 创建cache
func NewStandardCaches(c container.IContainer) *StandardCaches {
	return &StandardCaches{c: c}
}

//GetCaches GetCaches
func (s *StandardCaches) GetCache(name string) (q ICache, err error) {
	obj, err := s.c.GetOrCreate(cacheTypeNode, name, func(setting config.Config) (interface{}, error) {
		return NewCache(setting)
	})
	if err != nil {
		return nil, err
	}
	return obj.(ICache), nil
}
