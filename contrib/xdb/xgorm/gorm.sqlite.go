package xgorm

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/sqlite"
)

func init() {

	resolver := &sqliteResolver{Proto: "grom.sqlite"}
	xdb.Register(resolver)
	tpl.Register(tpl.NewFixed(resolver.Proto, "?"))
	callbackCache[resolver.Proto] = sqlite.Open

	rresolver := &sqliteResolver{Proto: "gorm.sqlite"}
	xdb.Register(rresolver)
	tpl.Register(tpl.NewFixed(rresolver.Proto, "?"))
	callbackCache[rresolver.Proto] = sqlite.Open
}

type sqliteResolver struct {
	Proto string
}

func (s *sqliteResolver) Name() string {
	return s.Proto
}

func (s *sqliteResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.ScanTo(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(s.Proto, cfg, opts...)
	if err != nil {
		return nil, err
	}
	tpl, err := tpl.GetDBTemplate(s.Proto)
	if err != nil {
		return nil, err
	}
	return &dbWrap{
		tpl:    tpl,
		proto:  s.Proto,
		gromDB: gromDB,
	}, nil
}
