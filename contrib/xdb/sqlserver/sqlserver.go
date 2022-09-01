package sqlserver

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

const Proto = "sqlserver"

type sqlserverResolver struct {
}

func (s *sqlserverResolver) Name() string {
	return Proto
}

func (s *sqlserverResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg.Conn, cfg.MaxOpen, cfg.MaxIdle, cfg.LifeTime)
}

func init() {
	xdb.Register(&sqlserverResolver{})
	tpl.Register(New(Proto, "@p"))
}
