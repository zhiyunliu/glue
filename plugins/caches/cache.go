package caches

import (
	"fmt"
	"sync"
	"time"
)

type ICache interface {
	String() string
	Get(key string) (string, error)
	Set(key string, val interface{}, expire int) error
	Del(key string) error
	HashGet(hk, key string) (string, error)
	HashDel(hk, key string) error
	Increase(key string) error
	Decrease(key string) error
	Expire(key string, dur time.Duration) error
	GetImpl() interface{}
}

var cacheMap sync.Map

func Registry(cache ICache) {
	if cache == nil {
		return
	}
	cacheMap.Store(cache.String(), cache)
}

func Get(key string) ICache {
	val, ok := cacheMap.Load(key)
	if !ok {
		panic(fmt.Errorf("不存在key=%s的cache实现", key))
	}
	return val.(ICache)
}
