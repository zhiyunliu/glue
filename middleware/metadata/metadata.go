package metadata

import (
	"strings"

	"github.com/zhiyunliu/velocity/context"

	"github.com/zhiyunliu/velocity/metadata"
	"github.com/zhiyunliu/velocity/middleware"
	"github.com/zhiyunliu/velocity/transport"
)

// Option is metadata option.
type Option func(*options)

type options struct {
	prefix []string
	md     metadata.Metadata
}

func (o *options) hasPrefix(key string) bool {
	k := strings.ToLower(key)
	for _, prefix := range o.prefix {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// WithConstants with constant metadata key value.
func WithConstants(md metadata.Metadata) Option {
	return func(o *options) {
		o.md = md
	}
}

// WithPropagatedPrefix with propagated key prefix.
func WithPropagatedPrefix(prefix ...string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// Server is middleware server-side metadata.
func Server(opts ...Option) middleware.Middleware {
	options := &options{
		prefix: []string{"x-md-"}, // x-md-global-, x-md-local
	}
	for _, o := range opts {
		o(options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			if tr, ok := transport.FromServerContext(ctx.Context()); ok {
				md := options.md.Clone()
				header := tr.RequestHeader()
				for _, k := range header.Keys() {
					if options.hasPrefix(k) {
						md.Set(k, header.Get(k))
					}
				}
				nctx := metadata.NewServerContext(ctx.Context(), md)
				ctx.ResetContext(nctx)
			}
			return handler(ctx)
		}
	}
}
