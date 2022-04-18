package sqlserver

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/zhiyunliu/gel/config"
	contribxdb "github.com/zhiyunliu/gel/contrib/xdb"
	"github.com/zhiyunliu/gel/xdb"
)

const Proto = "sqlserver"

type sqlserverResolver struct {
}

func (s *sqlserverResolver) Name() string {
	return Proto
}

func (s *sqlserverResolver) Resolve(setting config.Config) (xdb.IDB, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg.Conn, cfg.MaxOpen, cfg.MaxIdle, cfg.LifeTime)
}
func init() {
	xdb.Register(&sqlserverResolver{})
}