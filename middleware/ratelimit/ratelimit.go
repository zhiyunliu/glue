package ratelimit

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/standard"

	"github.com/zhiyunliu/glue/ratelimit"
	"github.com/zhiyunliu/glue/ratelimit/bbr"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/middleware"
)

// ErrLimitExceed is service unavailable due to rate limit exceeded.
var ErrLimitExceed = errors.New(429, "service unavailable due to rate limit exceeded")

// Option is ratelimit option.
type Option func(*options)

// WithLimiter set Limiter implementation,
// default is bbr limiter
func WithLimiter(limiter ratelimit.Limiter) Option {
	return func(o *options) {
		o.limiter = limiter
	}
}

type options struct {
	limiter ratelimit.Limiter
}

// Server ratelimiter middleware
func Server(opts ...Option) middleware.Middleware {
	options := &options{
		limiter: bbr.NewLimiter(),
	}
	for _, o := range opts {
		o(options)
	}
	return serverByOption(options)
}

func serverByConfig(cfg *Config) middleware.Middleware {
	std := standard.GetInstance(ratelimit.TypeNode).(ratelimit.Standard)
	provider := std.GetProvider(cfg.Limiter)

	options := &options{
		limiter: provider.Limiter(),
	}
	return serverByOption(options)
}

// Server ratelimiter middleware
func serverByOption(options *options) middleware.Middleware {

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			done, e := options.limiter.Allow()
			if e != nil {
				// rejected
				return ErrLimitExceed
			}
			// allowed
			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}

			done(ratelimit.DoneInfo{Err: err})
			return
		}
	}
}
