package xdb

import (
	"fmt"
	"sync"

	"github.com/zhiyunliu/gel/global"
)

var cache sync.Map
var prefix = "db_config_%s"

func init() {
	cache = sync.Map{}
}

func getCacheKey(name string) string {
	return fmt.Sprintf(prefix, name)
}

func GetConfig(name string) *Config {

	cfg := global.Config.Get("db")

	cfgVal := cfg.Value(name)
	dbCfg := &Config{}
	if err := cfgVal.Scan(dbCfg); err != nil {
		panic(fmt.Errorf("db.%s 读取错误:%w", name, err))
	}
	return dbCfg
}

func GetDB(name string) IDB {
	key := getCacheKey(name)
	obj, ok := cache.Load(key)
	if !ok {
		dbcfg := GetConfig(name)
		instance, err := NewDB(dbcfg.Proto, dbcfg.Conn, dbcfg.MaxOpen, dbcfg.MaxIdle, dbcfg.LifeTime)
		if err != nil {
			panic(fmt.Errorf("创建数据库失败:%w,name=%s", err, name))
		}
		cache.Store(key, instance)
		obj = instance
	}
	return obj.(IDB)
}
