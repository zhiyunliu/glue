package oracle

import (
	"fmt"

	_ "github.com/mattn/go-oci8"
	"github.com/zhiyunliu/gel/config"
	contribxdb "github.com/zhiyunliu/gel/contrib/xdb"
	"github.com/zhiyunliu/gel/xdb"
)

const Proto = "oracle"

type oracleResolver struct {
}

func (s *oracleResolver) Name() string {
	return Proto
}

func (s *oracleResolver) Resolve(setting config.Config) (xdb.IDB, error) {
	cfg := &contribxdb.Config{}
	err := setting.Scan(cfg)
	if err != nil {
		return nil, fmt.Errorf("读取DB配置:%w", err)
	}
	return contribxdb.NewDB(Proto, cfg.Conn, cfg.MaxOpen, cfg.MaxIdle, cfg.LifeTime)
}
func init() {
	xdb.Register(&oracleResolver{})
}
