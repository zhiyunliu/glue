package redis

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	contribredis "github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/dlocker"
)

func TestAcquireWaitsUntilReleased(t *testing.T) {
	client, server := newTestRedis(t)
	owner := client.Build("wait-release", dlocker.WithData("owner"))
	waiter := client.Build("wait-release", dlocker.WithData("waiter"), dlocker.WithTimeout(2), dlocker.WithFencing()).(*Lock)
	acquired, err := owner.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)

	result := make(chan error, 1)
	go func() {
		ok, err := waiter.Acquire(context.Background(), 3)
		if err == nil && !ok {
			err = errors.New("未获取到锁")
		}
		result <- err
	}()
	select {
	case err := <-result:
		t.Fatalf("锁释放前等待已结束: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	released, err := owner.Release(context.Background())
	require.NoError(t, err)
	require.True(t, released)
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("释放锁后未能及时获取")
	}
	require.Equal(t, 3*time.Second, server.TTL(waiter.stateKeys()[0]))
	require.Equal(t, int64(1), waiter.holdCount.Load())
	require.NotZero(t, waiter.FencingToken())
}

func TestAcquireWaitTimeoutDoesNotChangeState(t *testing.T) {
	client, server := newTestRedis(t)
	owner := client.Build("wait-timeout", dlocker.WithData("owner"))
	waiter := client.Build("wait-timeout", dlocker.WithData("waiter"), dlocker.WithTimeout(1), dlocker.WithAutoRenewal(), dlocker.WithFencing()).(*Lock)
	acquired, err := owner.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)

	started := time.Now()
	acquired, err = waiter.Acquire(context.Background(), 3)
	require.False(t, acquired)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.GreaterOrEqual(t, time.Since(started), time.Second)
	require.Less(t, time.Since(started), 2*time.Second)
	require.Zero(t, waiter.holdCount.Load())
	require.Zero(t, waiter.FencingToken())
	require.False(t, waiter.state.Load())
	value, err := server.Get(waiter.stateKeys()[0])
	require.NoError(t, err)
	require.Equal(t, "owner", value)
	value, err = server.Get(waiter.stateKeys()[1])
	require.NoError(t, err)
	require.Equal(t, "1", value)
}

func TestAcquireWaitRespectsParentContext(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(map[bool]string{false: "截止时间", true: "主动取消"}[canceled], func(t *testing.T) {
			client, _ := newTestRedis(t)
			owner := client.Build("parent", dlocker.WithData("owner"))
			waiter := client.Build("parent", dlocker.WithData("waiter"), dlocker.WithTimeout(3))
			acquired, err := owner.Acquire(context.Background(), 10)
			require.NoError(t, err)
			require.True(t, acquired)
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
			defer cancel()
			expected := context.DeadlineExceeded
			if canceled {
				timer := time.AfterFunc(60*time.Millisecond, cancel)
				defer timer.Stop()
				expected = context.Canceled
			}
			started := time.Now()
			acquired, err = waiter.Acquire(ctx, 10)
			require.False(t, acquired)
			require.ErrorIs(t, err, expected)
			require.Less(t, time.Since(started), time.Second)
		})
	}
}

func TestAcquireWithoutTimeoutOnlyAttemptsOnce(t *testing.T) {
	for _, opts := range [][]dlocker.Option{nil, {dlocker.WithTimeout(0)}} {
		client, _ := newTestRedis(t)
		owner := client.Build("single", dlocker.WithData("owner"))
		waiter := client.Build("single", append(opts, dlocker.WithData("waiter"))...)
		acquired, err := owner.Acquire(context.Background(), 10)
		require.NoError(t, err)
		require.True(t, acquired)
		started := time.Now()
		acquired, err = waiter.Acquire(context.Background(), 10)
		require.NoError(t, err)
		require.False(t, acquired)
		require.Less(t, time.Since(started), 500*time.Millisecond)
	}
}

func TestAcquireWaitSucceedsAfterLeaseExpires(t *testing.T) {
	client, server := newTestRedis(t)
	owner := client.Build("expired", dlocker.WithData("owner"))
	waiter := client.Build("expired", dlocker.WithData("waiter"), dlocker.WithTimeout(2))
	acquired, err := owner.Acquire(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, acquired)
	timer := time.AfterFunc(100*time.Millisecond, func() { server.FastForward(time.Second) })
	defer timer.Stop()
	acquired, err = waiter.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
}

func TestAcquireTimeoutPreservesReentrantBehavior(t *testing.T) {
	client, _ := newTestRedis(t)
	lock := client.Build("reentrant-timeout", dlocker.WithTimeout(1)).(*Lock)
	for range 2 {
		acquired, err := lock.Acquire(context.Background(), 10)
		require.NoError(t, err)
		require.True(t, acquired)
	}
	require.Equal(t, int64(2), lock.holdCount.Load())
	for range 2 {
		released, err := lock.Release(context.Background())
		require.NoError(t, err)
		require.True(t, released)
	}
}

func TestAcquireTimeoutRejectsNonReentrantOwnership(t *testing.T) {
	client, _ := newTestRedis(t)
	lock := client.Build("non-reentrant-timeout", dlocker.WithTimeout(1), dlocker.WithReentrant(false)).(*Lock)
	acquired, err := lock.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	acquired, err = lock.Acquire(ctx, 10)
	require.False(t, acquired)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, int64(1), lock.holdCount.Load())
}

func TestAcquireRejectsNegativeTimeout(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("negative", dlocker.WithTimeout(-1)).(*Lock)
	acquired, err := lock.Acquire(context.Background(), 10)
	require.False(t, acquired)
	require.Error(t, err)
	require.False(t, server.Exists(lock.stateKeys()[0]))
}

func TestAcquireCanceledContextDoesNotAcquire(t *testing.T) {
	client, server := newTestRedis(t)
	for _, timeout := range []int{0, 1} {
		lock := client.Build("canceled", dlocker.WithTimeout(timeout)).(*Lock)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		acquired, err := lock.Acquire(ctx, 10)
		require.False(t, acquired)
		require.ErrorIs(t, err, context.Canceled)
		require.False(t, server.Exists(lock.stateKeys()[0]))
	}
}

func TestAcquireWaitReturnsRedisErrorImmediately(t *testing.T) {
	client, server := newTestRedis(t)
	server.SetError("ERR 加锁失败")
	lock := client.Build("error", dlocker.WithTimeout(3))
	started := time.Now()
	acquired, err := lock.Acquire(context.Background(), 10)
	require.False(t, acquired)
	require.ErrorContains(t, err, "加锁失败")
	require.Less(t, time.Since(started), time.Second)
}

func TestAcquireTimeoutDoesNotStopAutoRenewal(t *testing.T) {
	client, _ := newTestRedis(t)
	lock := client.Build("renew-after-acquire", dlocker.WithTimeout(1), dlocker.WithAutoRenewal()).(*Lock)
	acquired, err := lock.Acquire(context.Background(), 2)
	require.NoError(t, err)
	require.True(t, acquired)
	t.Cleanup(func() { lock.stopAutoRenewal(0, nil) })
	time.Sleep(1100 * time.Millisecond)
	require.True(t, lock.state.Load())
	require.NoError(t, lock.Renewal(context.Background(), 2))
}

// 丢弃加锁请求以模拟 Redis 无响应，其他请求正常传输。
type stalledEvalConn struct {
	net.Conn
}

func (c stalledEvalConn) Write(p []byte) (int, error) {
	if strings.Contains(string(p), "\r\neval\r\n") {
		return len(p), nil
	}
	return c.Conn.Write(p)
}

func TestAcquireWaitBoundsRedisRead(t *testing.T) {
	_, server := newTestRedis(t)
	client := goredis.NewClient(&goredis.Options{
		Addr: server.Addr(), ContextTimeoutEnabled: true, MaxRetries: -1,
		ReadTimeout: 5 * time.Second,
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			return stalledEvalConn{conn}, nil
		},
	})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	lock := (&Redis{client: &contribredis.Client{UniversalClient: client}}).Build("stalled", dlocker.WithTimeout(1))
	started := time.Now()
	acquired, err := lock.Acquire(context.Background(), 10)
	require.False(t, acquired)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(started), 2*time.Second)
}
