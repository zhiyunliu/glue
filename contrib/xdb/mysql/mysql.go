package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/expression"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

const Proto = "mysql"

type mysqlResolver struct {
}

func (s *mysqlResolver) Name() string {
	return Proto
}

func (s *mysqlResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.ScanTo(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg, opts...)
}
func init() {
	tplMatcher := xdb.NewTemplateMatcher(expression.DefaultExpressionMatchers...)
	tplstmtProcessor := xdb.NewStmtDbTypeProcessor()

	xdb.Register(&mysqlResolver{})
	xdb.RegistTemplate(tpl.NewFixed(Proto, "?", tplMatcher, tplstmtProcessor))

}
