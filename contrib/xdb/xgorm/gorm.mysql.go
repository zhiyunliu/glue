package xgorm

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/mysql"
)

func init() {

	resolver := &mysqlResolver{Proto: "grom.mysql"}
	xdb.Register(resolver)
	tpl.Register(tpl.NewFixed(resolver.Proto, "?"))
	callbackCache[resolver.Proto] = mysql.Open

	rresolver := &mysqlResolver{Proto: "gorm.mysql"}
	xdb.Register(rresolver)
	tpl.Register(tpl.NewFixed(rresolver.Proto, "?"))
	callbackCache[rresolver.Proto] = mysql.Open

}

type mysqlResolver struct {
	Proto string
}

func (s *mysqlResolver) Name() string {
	return s.Proto
}

func (s *mysqlResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(s.Proto, cfg)
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
