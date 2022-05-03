package sqlite

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zhiyunliu/gel/config"
	contribxdb "github.com/zhiyunliu/gel/contrib/xdb"
	"github.com/zhiyunliu/gel/contrib/xdb/tpl"
	"github.com/zhiyunliu/gel/xdb"
)

const Proto = "sqlite"

type sqliteResolver struct {
}

func (s *sqliteResolver) Name() string {
	return Proto
}

func (s *sqliteResolver) Resolve(setting config.Config) (xdb.IDB, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg.Conn, cfg.MaxOpen, cfg.MaxIdle, cfg.LifeTime)
}
func init() {
	xdb.Register(&sqliteResolver{})
	tpl.Register(tpl.NewFixed(Proto, "?"))
}
