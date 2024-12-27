package xgorm

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/expression"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/driver/mysql"
)

func init() {
	tplMatcher := xdb.NewTemplateMatcher(expression.DefaultExpressionMatchers...)
	tplstmtProcessor := xdb.NewStmtDbTypeProcessor()

	resolver := &mysqlResolver{Proto: "grom.mysql"}
	xdb.Register(resolver)
	xdb.RegistTemplate(tpl.NewFixed(resolver.Proto, "?", tplMatcher, tplstmtProcessor))
	callbackCache[resolver.Proto] = mysql.Open

	rresolver := &mysqlResolver{Proto: "gorm.mysql"}
	xdb.Register(rresolver)
	xdb.RegistTemplate(tpl.NewFixed(rresolver.Proto, "?", tplMatcher, tplstmtProcessor))
	callbackCache[rresolver.Proto] = mysql.Open

}

type mysqlResolver struct {
	Proto string
}

func (s *mysqlResolver) Name() string {
	return s.Proto
}

func (s *mysqlResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.ScanTo(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	gromDB, err := buildGormDB(s.Proto, cfg, opts...)
	if err != nil {
		return nil, err
	}
	tpl, err := xdb.GetTemplate(s.Proto)
	if err != nil {
		return nil, err
	}
	return &dbWrap{
		tpl:    tpl,
		proto:  s.Proto,
		gromDB: gromDB,
	}, nil
}
