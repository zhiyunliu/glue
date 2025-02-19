package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/golibs/xrandom"
	"golang.org/x/sync/errgroup"
)

const (
	lockCommand = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	if ARGV[3] == "false" then
		return "NOK"
	end 
    redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    return "OK"
else
    return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
end`
	delCommand = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`

	leaseCommand = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
    return 0
end`

	//randomLen = 16
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
	rndVal      string
	opts        *dlocker.Options
	state       atomic.Bool
	releaseChan chan struct{}
	group       errgroup.Group
}

// NewLock returns a Lock.
func newLock(client *Redis, key string, opts *dlocker.Options) *Lock {
	var rndval string
	if opts.Data != "" {
		rndval = opts.Data
	} else {
		rndval = xrandom.Str(16)
	}
	return &Lock{
		client:      client,
		key:         key,
		rndVal:      rndval,
		opts:        opts,
		releaseChan: make(chan struct{}, 1),
		group:       errgroup.Group{},
	}
}

// Acquire acquires the lock.
// 单位：秒
// 加锁
func (rl *Lock) Acquire(ctx context.Context, expire int) (bool, error) {
	if expire <= 0 {
		return false, fmt.Errorf("expire 参数必须大于0")
	}
	// 获取过期时间
	// 默认锁过期时间为500ms，防止死锁
	resp, err := rl.client.Eval(ctx,
		lockCommand,
		[]string{rl.key},
		[]string{
			rl.rndVal,
			strconv.Itoa(expire*1000 + tolerance), //换算成毫秒
			strconv.FormatBool(rl.opts.Reentrant),
		},
	)
	if err == goredis.Nil {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error on acquiring lock for %s, %s", rl.key, err.Error())
	} else if resp == nil {
		return false, nil
	}

	reply, ok := resp.(string)
	if ok && strings.EqualFold(reply, "OK") {
		if rl.opts.AutoRenewal {
			_ = rl.autoRenewalCallback(expire)
		}
		return true, nil
	}
	return false, nil
}

// Release releases the lock.
// 释放锁
func (rl *Lock) Release(ctx context.Context) (bool, error) {
	resp, err := rl.client.Eval(ctx, delCommand, []string{rl.key}, []string{rl.rndVal})
	if err != nil {
		return false, err
	}

	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	succ := reply == 1
	if succ && rl.opts.AutoRenewal {
		old := rl.state.Load()
		if !old {
			return succ, nil
		}
		//确认没有变动
		if !rl.state.CompareAndSwap(old, false) {
			return succ, nil
		}

		select {
		case rl.releaseChan <- struct{}{}:
		default:
		}
	}

	return succ, nil
}

// 单位：秒
// 续约
func (rl *Lock) Renewal(ctx context.Context, expire int) error {
	resp, err := rl.client.Eval(ctx, leaseCommand, []string{rl.key}, []string{
		rl.rndVal,
		strconv.Itoa(expire*1000 + tolerance),
	})
	if err != nil {
		return fmt.Errorf("expire %+v,err:%+v", resp, err)
	}

	_, ok := resp.(int64)
	if !ok {
		return fmt.Errorf("expire %+v", resp)
	}

	return nil
}

func (rl *Lock) autoRenewalCallback(expire int) error {
	old := rl.state.Load()
	//原值已经在锁定中
	if old {
		return nil
	}

	//确认没有变动
	if !rl.state.CompareAndSwap(old, true) {
		return nil
	}
	autoRenewalTick := expire / 2
	if autoRenewalTick < 1 {
		autoRenewalTick = 1
	}

	rl.group.Go(func() error {
		ticker := time.NewTicker(time.Second * time.Duration(autoRenewalTick))
		for {
			select {
			case <-ticker.C:
				_ = rl.Renewal(context.Background(), expire)
			case <-rl.releaseChan:
				rl.state.Store(false)
				return nil
			}
		}
	})
	return nil
}
