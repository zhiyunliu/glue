package metrics

import (
	"strconv"
	"time"

	"github.com/zhiyunliu/velocity/context"

	"github.com/zhiyunliu/velocity/errors"
	"github.com/zhiyunliu/velocity/metrics"
	"github.com/zhiyunliu/velocity/middleware"
)

// Option is metrics option.
type Option func(*options)

// WithRequests with requests counter.
func WithRequests(c metrics.Counter) Option {
	return func(o *options) {
		o.requests = c
	}
}

// WithSeconds with seconds histogram.
func WithSeconds(c metrics.Observer) Option {
	return func(o *options) {
		o.seconds = c
	}
}

type options struct {
	// counter: <client/server>_requests_code_total{kind, operation, code, reason}
	requests metrics.Counter
	// histogram: <client/server>_requests_seconds_bucket{kind, operation}
	seconds metrics.Observer
}

// Server is middleware server-side metrics.
func Server(opts ...Option) middleware.Middleware {
	op := options{}
	for _, o := range opts {
		o(&op)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			var (
				code      int
				reason    string
				kind      string = ctx.ServerType()
				operation string = ctx.Request().Path().FullPath()
			)
			startTime := time.Now()

			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}

			if se := errors.FromError(err); se != nil {
				code = int(se.Code)
			}
			if op.requests != nil {
				op.requests.With(kind, operation, strconv.Itoa(code), reason).Inc()
			}
			if op.seconds != nil {
				op.seconds.With(kind, operation).Observe(time.Since(startTime).Seconds())
			}
			return reply
		}
	}
}
