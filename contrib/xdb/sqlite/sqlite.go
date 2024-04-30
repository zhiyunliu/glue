package sqlite

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

const Proto = "sqlite"

type sqliteResolver struct {
}

func (s *sqliteResolver) Name() string {
	return Proto
}

func (s *sqliteResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.ScanTo(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg, opts...)
}
func init() {
	xdb.Register(&sqliteResolver{})
	tpl.Register(tpl.NewFixed(Proto, "?"))
}
