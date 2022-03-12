package caches

import (
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/container"
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
func (s *StandardCaches) GetCaches(name string) (q ICache, err error) {
	obj, err := s.c.GetOrCreate(cacheTypeNode, name, func(setting *config.Setting) (interface{}, error) {
		return NewCache(setting)
	})
	if err != nil {
		return nil, err
	}
	return obj.(ICache), nil
}
