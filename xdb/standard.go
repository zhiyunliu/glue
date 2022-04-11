package xdb

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
)

const dbTypeNode = "dbs"

//StandardDB
type StandardDB struct {
	container container.Container
}

//NewStandardDBs 创建DB
func NewStandardDB(container container.Container) *StandardDB {
	return &StandardDB{container: container}
}

//GetDB
func (s *StandardDB) GetDB(name string) (q IDB) {
	obj, err := s.container.GetOrCreate(dbTypeNode, name, func(cfg config.Config) (interface{}, error) {
		dbcfg := cfg.Get(dbTypeNode).Get(name)
		return newDB(dbcfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(IDB)
}
