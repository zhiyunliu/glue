package sqlserver

import (
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

const Proto = "sqlserver"
const ArgumentPrefix = "p_"

type sqlserverResolver struct {
}

func (s *sqlserverResolver) Name() string {
	return Proto
}

func (s *sqlserverResolver) Resolve(setting config.Config) (interface{}, error) {
	cfg := contribxdb.NewConfig()
	err := setting.Scan(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg)
}

func init() {
	xdb.Register(&sqlserverResolver{})
	tpl.Register(New(Proto, ArgumentPrefix))
}
