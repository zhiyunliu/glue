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

func (s *sqlserverResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.ScanTo(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置(%s):%w", connName, err)
	}
	return contribxdb.NewDB(Proto, cfg, opts...)
}

func init() {
	xdb.Register(&sqlserverResolver{})
	tpl.Register(New(Proto, ArgumentPrefix))
}
