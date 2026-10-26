package redis

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestLockClientContextAndRetryOverrides(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := newRedis("lock", &Options{
		Addrs: []string{server.Addr()}, MaxRetries: 5,
	}, map[string]any{
		"context_timeout_enabled": true,
		"max_retries":             -1,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	opts := client.UniversalClient.(*goredis.Client).Options()
	require.True(t, opts.ContextTimeoutEnabled)
	require.Zero(t, opts.MaxRetries)
}

func TestRedisDefaultsPreserveContextAndRetryBehavior(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := NewByOpts("defaults", WithAddrs(server.Addr()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	opts := client.UniversalClient.(*goredis.Client).Options()
	require.False(t, opts.ContextTimeoutEnabled)
	require.Equal(t, 3, opts.MaxRetries)
}
