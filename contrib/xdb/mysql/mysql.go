package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

const Proto = "mysql"

type mysqlResolver struct {
}

func (s *mysqlResolver) Name() string {
	return Proto
}

func (s *mysqlResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg.Conn, cfg.MaxOpen, cfg.MaxIdle, cfg.LifeTime)
}
func init() {
	xdb.Register(&mysqlResolver{})
	tpl.Register(tpl.NewFixed(Proto, "?"))

}
