package xgorm

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/postgres"
)

const postgresProto = "grom.postgres"

func init() {
	xdb.Register(&postgresResolver{})
	tpl.Register(tpl.NewFixed(postgresProto, "$"))
	callbackCache[postgresProto] = postgres.Open
}

type postgresResolver struct {
}

func (s *postgresResolver) Name() string {
	return postgresProto
}

func (s *postgresResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(postgresProto, cfg)
	if err != nil {
		return nil, err
	}
	tpl, err := tpl.GetDBTemplate(postgresProto)
	if err != nil {
		return nil, err
	}
	return &dbWrap{
		tpl:    tpl,
		gromDB: gromDB,
	}, nil
}
