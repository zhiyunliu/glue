package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/standard"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/metrics"
	"github.com/zhiyunliu/glue/middleware"
)

// Option is metrics option.
type Option func(*options)

// WithCounter with requests counter.
func WithCounter(c metrics.Counter) Option {
	return func(o *options) {
		o.counter = c
	}
}

// WithObserver
func WithObserver(c metrics.Observer) Option {
	return func(o *options) {
		o.observer = c
	}
}

// WithGauge
func WithGauge(c metrics.Gauge) Option {
	return func(o *options) {
		o.gauge = c
	}
}

type options struct {
	counter  metrics.Counter
	observer metrics.Observer
	gauge    metrics.Gauge
}

func Server(opts ...Option) middleware.Middleware {
	op := options{}
	for _, o := range opts {
		o(&op)
	}
	return serverByOptions(&op)
}

func serverByConfig(cfg *Config) middleware.Middleware {
	op := &options{}

	stdMetric := standard.GetInstance(metrics.TypeNode).(metrics.StandardMetric)
	provider := stdMetric.GetProvider(cfg.Proto)

	op.counter = provider.Counter()
	op.observer = provider.Observer()

	return serverByOptions(op)
}

// Server is middleware server-side metrics.
func serverByOptions(op *options) middleware.Middleware {

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			var (
				code   int = http.StatusOK
				reason string
				kind   string = ctx.ServerType()
				path   string = ctx.Request().Path().FullPath()
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
			if op.counter != nil {
				op.counter.With(kind, path, strconv.Itoa(code), reason).Inc()
			}
			if op.observer != nil {
				op.observer.With(kind, path).Observe(time.Since(startTime).Seconds())
			}
			return reply
		}
	}
}
