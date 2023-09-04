package xdb

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

var (
	DateFormat = "2006-01-02 15:04:05"
)

const DbTypeNode = "dbs"
const _defaultName = "default"

// StandardDB
type StandardDB interface {
	GetDB(name ...string) (q IDB)
	GetImpl(name ...string) interface{}
}

// StandardDB
type xDB struct {
	container container.Container
}

// NewStandardDBs 创建DB
func NewStandardDB(container container.Container) StandardDB {
	return &xDB{container: container}
}

func (s *xDB) GetImpl(name ...string) interface{} {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}
	if realName == "" {
		realName = _defaultName
	}
	obj, err := s.container.GetOrCreate(DbTypeNode, realName, func(cfg config.Config) (interface{}, error) {
		dbcfg := cfg.Get(DbTypeNode).Get(realName)
		return newDB(realName, dbcfg)
	})
	if err != nil {
		panic(err)
	}
	return obj
}

// GetDB
func (s *xDB) GetDB(name ...string) (q IDB) {
	obj := s.GetImpl(name...)
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
