package xgorm

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/sqlserver"
)

const mssqlProto = "grom.mssql"

func init() {
	xdb.Register(&mssqlResolver{})
	tpl.Register(tpl.NewFixed(mssqlProto, "?"))
	callbackCache[mssqlProto] = sqlserver.Open
}

type mssqlResolver struct {
}

func (s *mssqlResolver) Name() string {
	return mssqlProto
}

func (s *mssqlResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(mssqlProto, cfg)
	if err != nil {
		return nil, err
	}
	tpl, err := tpl.GetDBTemplate(mssqlProto)
	if err != nil {
		return nil, err
	}
	return &dbWrap{
		tpl:    tpl,
		gromDB: gromDB,
	}, nil
}
