package circuitbreaker

//todo ： circuitbreaker 需要放在client 端的请求中

import (
	"github.com/zhiyunliu/glue/context"

	"github.com/zhiyunliu/glue/circuitbreaker"
	"github.com/zhiyunliu/glue/circuitbreaker/sre"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/golibs/group"
)

// ErrNotAllowed is request failed due to circuit breaker triggered.
var ErrNotAllowed = errors.New(503, "request failed due to circuit breaker triggered")

// Option is circuit breaker option.
type Option func(*options)

// WithGroup with circuit breaker group.
// NOTE: implements generics circuitbreaker.CircuitBreaker
func WithGroup(g *group.Group) Option {
	return func(o *options) {
		o.group = g
	}
}

type options struct {
	group *group.Group
}

// Client circuitbreaker middleware will return errBreakerTriggered when the circuit
// breaker is triggered and the request is rejected directly.
func Client(opts ...Option) middleware.Middleware {
	opt := &options{
		group: group.NewGroup(func() interface{} {
			return sre.NewBreaker()
		}),
	}
	for _, o := range opts {
		o(opt)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {

			path := ctx.Request().Path().GetURL().Path
			breaker := opt.group.Get(path).(circuitbreaker.CircuitBreaker)
			if err := breaker.Allow(); err != nil {
				// rejected
				// NOTE: when client reject requets locally,
				// continue add counter let the drop ratio higher.
				breaker.MarkFailed()
				return ErrNotAllowed
			}
			// allowed
			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}

			if err != nil && errors.IsInternalServer(err) {
				breaker.MarkFailed()
			} else {
				breaker.MarkSuccess()
			}
			return reply
		}
	}
}
