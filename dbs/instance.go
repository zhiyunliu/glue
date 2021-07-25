package dbs

import (
	"fmt"
	"sync"
)

var cache sync.Map

func init() {
	cache = sync.Map{}
}

func SetDbConfig(name string, cfg *DB) {
	cache.Store(fmt.Sprintf("db_config_%s", name), cfg)
}

func GetDbConfig(name string) *Config {
	val, ok := cache.Load(fmt.Sprintf("db_config_%s", name))
	if !ok {
		panic(fmt.Errorf("不存在DB=%s的配置", name))
	}
	return val.(*Config)
}

func GetDB() IDB {
	return GetDBByName("db")
}

func GetDBByName(name string) IDB {
	key := fmt.Sprintf("db_instance_%s", name)

	obj, ok := cache.Load(key)
	if !ok {
		dbcfg := GetDbConfig(name)
		instance, err := NewDB(dbcfg.Provider, dbcfg.ConnString, dbcfg.MaxOpen, dbcfg.MaxIdle, dbcfg.LifeTime)
		if err != nil {
			panic(fmt.Errorf("创建数据库失败:%w", err))
		}
		cache.Store(key, instance)
		obj = instance
	}
	return obj.(IDB)
}
