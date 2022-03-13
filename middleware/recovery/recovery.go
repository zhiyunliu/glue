package recovery

import (
	"runtime"

	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/errors"

	"github.com/zhiyunliu/velocity/log"
	"github.com/zhiyunliu/velocity/middleware"
)

// ErrUnknownRequest is unknown request error.
var ErrUnknownRequest = errors.InternalServer("Recovery:unknown request error")

// HandlerFunc is recovery handler func.
type HandlerFunc func(ctx context.Context, err interface{}) error

// Option is recovery option.
type Option func(*options)

type options struct {
	handler HandlerFunc
}

// WithHandler with recovery handler.
func WithHandler(h HandlerFunc) Option {
	return func(o *options) {
		o.handler = h
	}
}

// Recovery is a server middleware that recovers from any panics.
func Recovery(opts ...Option) middleware.Middleware {
	op := options{
		handler: func(ctx context.Context, err interface{}) error {
			return ErrUnknownRequest
		},
	}
	for _, o := range opts {
		o(&op)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			defer func() {
				if rerr := recover(); rerr != nil {
					buf := make([]byte, 64<<10) //nolint:gomnd
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					ctx.Log().Logf(log.LevelError, "%v: \n%s\n", rerr, buf)

					reply = op.handler(ctx, rerr)
				}
			}()
			return handler(ctx)
		}
	}
}
