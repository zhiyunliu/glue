package xdb

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const DbTypeNode = "dbs"
const _defaultName = "default"

//StandardDB
type StandardDB interface {
	GetDB(name ...string) (q IDB)
}

//StandardDB
type xDB struct {
	container container.Container
}

//NewStandardDBs 创建DB
func NewStandardDB(container container.Container) StandardDB {
	return &xDB{container: container}
}

//GetDB
func (s *xDB) GetDB(name ...string) (q IDB) {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}

	obj, err := s.container.GetOrCreate(DbTypeNode, realName, func(cfg config.Config) (interface{}, error) {
		dbcfg := cfg.Get(DbTypeNode).Get(realName)
		return newDB(dbcfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(IDB)
}

type xBuilder struct{}

func NewBuilder() container.StandardBuilder {
	return &xBuilder{}
}

func (xBuilder) Name() string {
	return DbTypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardDB(c)
}
