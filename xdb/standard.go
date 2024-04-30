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
	GetDB(name string, opts ...Option) (q IDB)
	GetImpl(name string, opts ...Option) interface{}
}

// StandardDB
type xDB struct {
	container container.Container
}

// NewStandardDBs 创建DB
func NewStandardDB(container container.Container) StandardDB {
	return &xDB{container: container}
}

func (s *xDB) GetImpl(name string, opts ...Option) interface{} {
	realName := name
	if realName == "" {
		realName = _defaultName
	}

	obj, err := s.container.GetOrCreate(DbTypeNode, realName, func(cfg config.Config) (interface{}, error) {
		dbcfg := cfg.Get(DbTypeNode).Get(realName)
		return newDB(realName, dbcfg, opts...)
	}, s.getUniqueKey(opts...))
	if err != nil {
		panic(err)
	}
	return obj
}

// GetDB
func (s *xDB) GetDB(name string, opts ...Option) (q IDB) {
	obj := s.GetImpl(name, opts...)
	return obj.(IDB)
}

func (s *xDB) getUniqueKey(opts ...Option) string {
	if len(opts) == 0 {
		return ""
	}
	tmpCfg := &Config{}
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
	return DbTypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewStandardDB(c)
}
