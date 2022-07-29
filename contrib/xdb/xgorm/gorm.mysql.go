package xgorm

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/mysql"
)

const mysqlProto = "grom.mysql"

func init() {
	xdb.Register(&mysqlResolver{})
	tpl.Register(tpl.NewFixed(mysqlProto, "?"))
	callbackCache[mysqlProto] = mysql.Open

}

type mysqlResolver struct {
}

func (s *mysqlResolver) Name() string {
	return mysqlProto
}

func (s *mysqlResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(mysqlProto, cfg)
	if err != nil {
		return nil, err
	}
	tpl, err := tpl.GetDBTemplate(mysqlProto)
	if err != nil {
		return nil, err
	}
	return &dbWrap{
		tpl:    tpl,
		gromDB: gromDB,
	}, nil
}
