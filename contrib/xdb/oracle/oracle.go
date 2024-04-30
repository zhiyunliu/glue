package oracle

import (
	"fmt"

	_ "github.com/mattn/go-oci8"
	"github.com/zhiyunliu/glue/config"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

const Proto = "oracle"

type oracleResolver struct {
}

func (s *oracleResolver) Name() string {
	return Proto
}

func (s *oracleResolver) Resolve(connName string, setting config.Config, opts ...xdb.Option) (interface{}, error) {
	cfg := contribxdb.NewConfig(connName)
	err := setting.ScanTo(cfg.Cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg, opts...)
}
func init() {
	xdb.Register(&oracleResolver{})
	tpl.Register(tpl.NewSeq(Proto, ":"))

}
