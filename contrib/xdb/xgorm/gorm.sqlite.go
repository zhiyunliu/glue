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

const sqliteProto = "grom.sqlite"

func init() {
	xdb.Register(&sqliteResolver{})
	tpl.Register(tpl.NewFixed(sqliteProto, "?"))
	callbackCache[sqliteProto] = sqlite.Open
}

type sqliteResolver struct {
}

func (s *sqliteResolver) Name() string {
	return sqliteProto
}

func (s *sqliteResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(sqliteProto, cfg)
	if err != nil {
		return nil, err
	}
	tpl, err := tpl.GetDBTemplate(sqliteProto)
	if err != nil {
		return nil, err
	}
	return &dbWrap{
		tpl:    tpl,
		gromDB: gromDB,
	}, nil
}
