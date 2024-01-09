package xgorm

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/postgres"
)

func init() {

	resolver := &postgresResolver{Proto: "grom.postgres"}
	xdb.Register(resolver)
	tpl.Register(tpl.NewFixed(resolver.Proto, "$"))
	callbackCache[resolver.Proto] = postgres.Open

	rresolver := &postgresResolver{Proto: "gorm.postgres"}
	xdb.Register(rresolver)
	tpl.Register(tpl.NewFixed(rresolver.Proto, "$"))
	callbackCache[rresolver.Proto] = postgres.Open
}

type postgresResolver struct {
	Proto string
}

func (s *postgresResolver) Name() string {
	return s.Proto
}

func (s *postgresResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.Scan(cfg.Cfg)
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
		gromDB: gromDB,
	}, nil
}
