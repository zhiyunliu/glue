package redis

import (
	"fmt"
	"strconv"

	goredis "github.com/go-redis/redis/v7"
	"github.com/zhiyunliu/golibs/xrandom"
)

const (
	lockCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    return "OK"
else
    return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
end`
	delCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`
	leaseCommand = `return redis.call("EXPIRE", KEYS[1], ARGV[1])`

	randomLen = 16
	// 默认超时时间，防止死锁
	tolerance int = 500 // milliseconds
)

// A Lock is a redis lock.
type Lock struct {
	// redis客户端
	client *Redis
	// 锁key
	key string
	// 锁value，防止锁被别人获取到
	rndVal string
}

// NewLock returns a Lock.
func NewLock(client *Redis, key string) *Lock {
	return &Lock{
		client: client,
		key:    key,
		rndVal: xrandom.Str(randomLen),
	}
}

// Acquire acquires the lock.
// 单位：秒
// 加锁
func (rl *Lock) Acquire(expire int) (bool, error) {
	if expire <= 0 {
		return false, fmt.Errorf("expire 参数必须大于0")
	}
	expire = expire*1000 + tolerance //换算成毫秒
	// 获取过期时间
	// 默认锁过期时间为500ms，防止死锁
	resp, err := rl.client.Eval(lockCommand, []string{rl.key}, []string{
		rl.rndVal, strconv.Itoa(expire),
	})
	if err == goredis.Nil {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error on acquiring lock for %s, %s", rl.key, err.Error())
	} else if resp == nil {
		return false, nil
	}

	reply, ok := resp.(string)
	if ok && reply == "OK" {
		return true, nil
	}
	return false, nil
}

// Release releases the lock.
// 释放锁
func (rl *Lock) Release() (bool, error) {
	resp, err := rl.client.Eval(delCommand, []string{rl.key}, []string{rl.rndVal})
	if err != nil {
		return false, err
	}

	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return reply == 1, nil
}

// 单位：秒
// 续约
func (rl *Lock) Renewal(expire int) error {
	resp, err := rl.client.Eval(leaseCommand, []string{rl.key}, []string{strconv.Itoa(expire)})
	if err != nil {
		return err
	}

	_, ok := resp.(int64)
	if !ok {
		return fmt.Errorf("expire %+v", resp)
	}

	return nil
}
