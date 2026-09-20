package redis

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	contribredis "github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/dlocker"
)

func newTestRedis(t *testing.T) (*Redis, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return &Redis{client: &contribredis.Client{UniversalClient: client}}, server
}

func TestReentrantReleaseKeepsLockUntilFinalRelease(t *testing.T) {
	client, _ := newTestRedis(t)
	owner := client.Build("reentrant", dlocker.WithData("owner"))
	competitor := client.Build("reentrant", dlocker.WithData("competitor"))

	acquired, err := owner.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)

	acquired, err = owner.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)

	released, err := owner.Release(context.Background())
	require.NoError(t, err)
	require.True(t, released)

	acquired, err = competitor.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.False(t, acquired)

	released, err = owner.Release(context.Background())
	require.NoError(t, err)
	require.True(t, released)

	acquired, err = competitor.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
}

func TestRenewalReportsLostLock(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("{lost}:lock", dlocker.WithData("owner")).(*Lock)

	acquired, err := lock.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	server.Del(lock.stateKeys()[0])

	require.Error(t, lock.Renewal(context.Background(), 10))
}

func TestAcquireUsesExactExpiration(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("{ttl}:lock", dlocker.WithData("owner")).(*Lock)

	acquired, err := lock.Acquire(context.Background(), 2)
	require.NoError(t, err)
	require.True(t, acquired)
	require.Equal(t, 2*time.Second, server.TTL(lock.stateKeys()[0]))
}

func TestAutoRenewalStopsAfterLockIsLost(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("{auto-renewal}:lock", dlocker.WithData("owner"), dlocker.WithAutoRenewal()).(*Lock)

	acquired, err := lock.Acquire(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, acquired)
	require.Eventually(t, lock.state.Load, time.Second, 10*time.Millisecond)

	server.Del(lock.stateKeys()[0])
	require.Eventually(t, func() bool {
		return !lock.state.Load()
	}, 2*time.Second, 10*time.Millisecond)
	require.ErrorIs(t, <-lock.Lost(), dlocker.ErrLockLost)
}

func TestFencingTokenIncreasesForEachOwnership(t *testing.T) {
	client, _ := newTestRedis(t)
	first := client.Build("fencing", dlocker.WithData("first")).(*Lock)
	second := client.Build("fencing", dlocker.WithData("second")).(*Lock)

	acquired, err := first.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	firstToken := first.FencingToken()
	require.NotZero(t, firstToken)

	acquired, err = first.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	require.Equal(t, firstToken, first.FencingToken())

	released, err := first.Release(context.Background())
	require.NoError(t, err)
	require.True(t, released)
	released, err = first.Release(context.Background())
	require.NoError(t, err)
	require.True(t, released)

	acquired, err = second.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	require.Greater(t, second.FencingToken(), firstToken)
}

func TestAcquireRejectsReentryWhenDisabled(t *testing.T) {
	client, _ := newTestRedis(t)
	lock := client.Build("non-reentrant", dlocker.WithData("owner"), dlocker.WithReentrant(false))

	acquired, err := lock.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)

	acquired, err = lock.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.False(t, acquired)
}

func TestReleaseStopsAutoRenewalWhenOwnershipIsLost(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("{release-lost}:lock", dlocker.WithData("owner"), dlocker.WithAutoRenewal()).(*Lock)

	acquired, err := lock.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	server.Del(lock.stateKeys()[0])

	released, err := lock.Release(context.Background())
	require.NoError(t, err)
	require.False(t, released)
	require.Eventually(t, func() bool {
		return !lock.state.Load()
	}, time.Second, 10*time.Millisecond)
	require.ErrorIs(t, <-lock.Lost(), dlocker.ErrLockLost)
}

func TestLocksWithSameHashTagRemainIndependent(t *testing.T) {
	client, _ := newTestRedis(t)
	first := client.Build("{shared}:first", dlocker.WithData("owner"))
	second := client.Build("{shared}:second", dlocker.WithData("competitor"))

	acquired, err := first.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)

	acquired, err = second.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
}

func TestStateKeysKeepOwnerKeyAndShareRedisSlot(t *testing.T) {
	for _, key := range []string{"plain-key", "{shared}:first"} {
		keys := lockStateKeys(key)
		require.Equal(t, key, keys[0])
		require.Equal(t, redisSlot(keys[0]), redisSlot(keys[1]))
		require.Equal(t, redisSlot(keys[0]), redisSlot(keys[2]))
	}
}

func TestReleaseErrorStopsAutoRenewal(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("release-error", dlocker.WithData("owner"), dlocker.WithAutoRenewal()).(*Lock)

	acquired, err := lock.Acquire(context.Background(), 10)
	require.NoError(t, err)
	require.True(t, acquired)
	server.Close()

	released, err := lock.Release(context.Background())
	require.Error(t, err)
	require.False(t, released)
	require.Eventually(t, func() bool {
		return !lock.state.Load()
	}, time.Second, 10*time.Millisecond)
}

func TestOneSecondLeaseRenewsBeforeExpiration(t *testing.T) {
	client, server := newTestRedis(t)
	lock := client.Build("short-lease", dlocker.WithData("owner"), dlocker.WithAutoRenewal()).(*Lock)

	acquired, err := lock.Acquire(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, acquired)
	server.FastForward(750 * time.Millisecond)
	time.Sleep(600 * time.Millisecond)

	require.Greater(t, server.TTL(lock.stateKeys()[0]), 500*time.Millisecond)
}
