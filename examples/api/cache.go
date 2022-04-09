package main

import (
	"time"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
)

type Cachedemo struct{}

func NewCache() *Cachedemo {
	return &Cachedemo{}
}

func (d *Cachedemo) GetHandle(ctx context.Context) interface{} {
	cacheObj := gel.Cache().GetCache("default")
	cacheObj.Set("key", time.Now().Nanosecond(), 10)

	val, err := cacheObj.Get("key")
	return map[string]interface{}{
		"val": val,
		"err": err,
	}
}

// Get(key string) (string, error)
// Set(key string, val interface{}, expire int) error
// Del(key string) error
// HashGet(hk, key string) (string, error)
// HashDel(hk, key string) error
// Increase(key string) error
// Decrease(key string) error
// Expire(key string, expire int) error
// GetImpl() interface{}
