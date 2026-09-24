package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/golibs/xrandom"
)

const (
	lockKeyPrefix = "dlocker:"

	lockCommand = `

local owner = redis.call("GET", KEYS[1])
if owner == ARGV[1] then
	if ARGV[3] == "false" then
		return {0, 0}
	end
	local count = tonumber(redis.call("GET", KEYS[2])) or 1
	redis.call("SET", KEYS[2], count + 1, "PX", ARGV[2])
	redis.call("PEXPIRE", KEYS[1], ARGV[2])
	if ARGV[4] == "false" then
		return {1, 0}
	end
	local token = redis.call("GET", KEYS[3])
	if not token then
		token = redis.call("INCR", KEYS[3])
	end
	return {1, token}
end

local acquired = redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
if not acquired then
	return {0, 0}
end

redis.call("SET", KEYS[2], 1, "PX", ARGV[2])
if ARGV[4] == "false" then
	return {1, 0}
end
return {1, redis.call("INCR", KEYS[3])}`
	delCommand = `
if redis.call("GET", KEYS[1]) ~= ARGV[1] then
	return -1
end

local count = tonumber(redis.call("GET", KEYS[2])) or 1
if count > 1 then
	redis.call("DECR", KEYS[2])
	return count - 1
end

redis.call("DEL", KEYS[1], KEYS[2])
return 0`

	leaseCommand = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	redis.call("PEXPIRE", KEYS[1], ARGV[2])
	if redis.call("EXISTS", KEYS[2]) == 1 then
		redis.call("PEXPIRE", KEYS[2], ARGV[2])
	end
	return 1
end
return 0`
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
	holdCount   atomic.Int64
	fencing     atomic.Uint64
	renewalMu   sync.Mutex
	renewalSeq  uint64
	renewalStop context.CancelFunc
	lostChan    chan error
	keys        []string
}

// NewLock returns a Lock.
func newLock(client *Redis, key string, opts *dlocker.Options) *Lock {
	var rndval string
	if opts.Data != "" {
		rndval = opts.Data
	} else {
		rndval = xrandom.Str(16)
	}
	lock := &Lock{
		client:   client,
		key:      key,
		rndVal:   rndval,
		opts:     opts,
		lostChan: make(chan error, 1),
	}
	lock.keys = lockStateKeys(key)
	return lock
}

func lockStateKeys(key string) []string {
	lockKey := lockKeyPrefix + key
	digest := sha256.Sum256([]byte(key))
	keyID := hex.EncodeToString(digest[:8])
	tag := ""
	if start := strings.IndexByte(lockKey, '{'); start >= 0 {
		if end := strings.IndexByte(lockKey[start+1:], '}'); end > 0 {
			tag = lockKey[start+1 : start+1+end]
		}
	}
	if tag == "" {
		tag = findHashTag(redisSlot(lockKey), keyID)
	}
	prefix := lockKeyPrefix + "{" + tag + "}:" + keyID + ":"
	return []string{lockKey, prefix + "count", prefix + "fencing"}
}

func (rl *Lock) stateKeys() []string {
	return rl.keys
}

func findHashTag(slot uint16, prefix string) string {
	for idx := 0; ; idx++ {
		tag := prefix + ":" + strconv.Itoa(idx)
		if redisSlot(tag) == slot {
			return tag
		}
	}
}

func redisSlot(key string) uint16 {
	if start := strings.IndexByte(key, '{'); start >= 0 {
		if end := strings.IndexByte(key[start+1:], '}'); end > 0 {
			key = key[start+1 : start+1+end]
		}
	}
	var crc uint16
	for idx := range len(key) {
		crc ^= uint16(key[idx]) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc % 16384
}

// Acquire acquires the lock.
// 单位：秒
// 加锁
func (rl *Lock) Acquire(ctx context.Context, expire int) (bool, error) {
	if expire <= 0 {
		return false, fmt.Errorf("expire 参数必须大于0")
	}
	resp, err := rl.client.Eval(ctx,
		lockCommand,
		rl.stateKeys(),
		[]string{
			rl.rndVal,
			strconv.Itoa(expire * 1000),
			strconv.FormatBool(rl.opts.Reentrant),
			strconv.FormatBool(rl.opts.Fencing),
		},
	)
	if err == goredis.Nil {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error on acquiring lock for %s, %s", rl.key, err.Error())
	} else if resp == nil {
		return false, nil
	}

	reply, ok := resp.([]interface{})
	if ok && len(reply) == 2 && reply[0] == int64(1) {
		token, parseErr := strconv.ParseUint(fmt.Sprint(reply[1]), 10, 64)
		if parseErr != nil {
			return false, fmt.Errorf("parse fencing token for %s: %w", rl.key, parseErr)
		}
		rl.fencing.Store(token)
		rl.holdCount.Add(1)
		if rl.opts.AutoRenewal {
			rl.startAutoRenewal(expire)
		}
		return true, nil
	}
	return false, nil
}

// Release releases the lock.
// 释放锁
func (rl *Lock) Release(ctx context.Context) (bool, error) {
	keys := rl.stateKeys()
	resp, err := rl.client.Eval(ctx, delCommand, keys[:2], []string{rl.rndVal})
	if err != nil {
		rl.holdCount.Store(0)
		rl.stopAutoRenewal(0, err)
		return false, err
	}

	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	if reply < 0 {
		rl.holdCount.Store(0)
		rl.stopAutoRenewal(0, dlocker.ErrLockLost)
		return false, nil
	}

	if rl.holdCount.Load() > 0 {
		rl.holdCount.Add(-1)
	}
	if reply == 0 || rl.holdCount.Load() == 0 {
		rl.stopAutoRenewal(0, nil)
	}
	return true, nil
}

// 单位：秒
// 续约
func (rl *Lock) Renewal(ctx context.Context, expire int) error {
	if expire <= 0 {
		return errors.New("dlocker[redis]Renewal.expire 参数必须大于0")
	}
	keys := rl.stateKeys()
	resp, err := rl.client.Eval(ctx, leaseCommand, keys[:2], []string{
		rl.rndVal,
		strconv.Itoa(expire * 1000),
	})
	if err != nil {
		return fmt.Errorf("expire %+v,err:%+v", resp, err)
	}

	reply, ok := resp.(int64)
	if !ok {
		return fmt.Errorf("expire %+v", resp)
	}
	if reply != 1 {
		return dlocker.ErrLockLost
	}

	return nil
}

func (rl *Lock) FencingToken() uint64 {
	return rl.fencing.Load()
}

func (rl *Lock) Lost() <-chan error {
	return rl.lostChan
}

func (rl *Lock) startAutoRenewal(expire int) {
	rl.renewalMu.Lock()
	defer rl.renewalMu.Unlock()
	if rl.state.Load() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	rl.renewalSeq++
	seq := rl.renewalSeq
	rl.renewalStop = cancel
	rl.state.Store(true)
	autoRenewalTick := time.Duration(expire) * time.Second / 2
	if autoRenewalTick < 100*time.Millisecond {
		autoRenewalTick = 100 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(autoRenewalTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := rl.Renewal(ctx, expire); err != nil {
					rl.holdCount.Store(0)
					rl.stopAutoRenewal(seq, err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (rl *Lock) stopAutoRenewal(seq uint64, err error) {
	rl.renewalMu.Lock()
	defer rl.renewalMu.Unlock()
	if seq != 0 && seq != rl.renewalSeq {
		return
	}
	if rl.renewalStop != nil {
		rl.renewalStop()
		rl.renewalStop = nil
	}
	rl.state.Store(false)
	if err != nil && !errors.Is(err, context.Canceled) {
		select {
		case rl.lostChan <- err:
		default:
		}
	}
}
