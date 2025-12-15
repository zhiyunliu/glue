package xdb

import (
	"os"

	"github.com/zhiyunliu/golibs/xtransform"
)

// 数据库连接重构方法
var ConnRefactor func(connName string, cfg *Config) (newcfg *Config, err error)

func DefaultRefactor(connName string, cfg *Config) (newcfg *Config, err error) {
	newcfg = cfg
	if DecryptConn != nil {
		newcfg.Conn, err = DecryptConn(connName, cfg.Conn)
		if err != nil {
			return
		}
	}
	//优化数据库链接配置
	newcfg.Conn = xtransform.TranslateCallback(newcfg.Conn, func(argName string) string {
		val := os.Getenv(argName)
		if len(val) > 0 {
			return val
		}
		return argName
	})

	if ConnRefactor != nil {
		newcfg, err = ConnRefactor(connName, newcfg)
	}
	return
}
