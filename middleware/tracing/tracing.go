package tracing

import (
	"github.com/zhiyunliu/gel/context"

	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/transport"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option is tracing option.
type Option func(*options)

type options struct {
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

// WithPropagator with tracer propagator.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opts *options) {
		opts.propagator = propagator
	}
}

// WithTracerProvider with tracer provider.
// Deprecated: use otel.SetTracerProvider(provider) instead.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(opts *options) {
		opts.tracerProvider = provider
	}
}

// Server returns a new server middleware for OpenTelemetry.
func Server(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindServer, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			if _, ok := transport.FromServerContext(ctx.Context()); ok {
				sctx, span := tracer.Start(ctx.Context(), ctx.Request().Path().FullPath(), ctx.Request().Header())
				ctx.ResetContext(sctx)
				setServerSpan(ctx, span, ctx.Request())
				defer func() { tracer.End(ctx.Context(), span, reply) }()
			}
			return handler(ctx)
		}
	}
}
