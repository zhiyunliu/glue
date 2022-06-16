package tracing

import (
	sctx "context"

	"github.com/zhiyunliu/glue/context"

	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/transport"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option is tracing option.
type Option func(*options)

type options struct {
	tracerName     string
	SpanKind       trace.SpanKind
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

var (
	_defaultName = "gel.tracer"
)

// WithTracerName with tracer name.
func WithTracerName(tracerName string) Option {
	return func(opts *options) {
		opts.tracerName = tracerName
	}
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

// Server ratelimiter middleware
func Server(opts ...Option) middleware.Middleware {
	options := &options{
		SpanKind:   trace.SpanKindServer,
		tracerName: _defaultName,
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	for _, o := range opts {
		o(options)
	}
	return serverByOption(options)
}

func serverByConfig(cfg *Config) middleware.Middleware {
	options := &options{
		SpanKind:   cfg.SpanKind,
		tracerName: _defaultName,
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	return serverByOption(options)
}

// Server ratelimiter middleware
func serverByOption(options *options) middleware.Middleware {
	tracer := newTracerByOpts(options.SpanKind, options)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			if _, ok := transport.FromServerContext(ctx.Context()); ok {
				sctx, span := tracer.Start(ctx.Context(), ctx.Request().Path().FullPath(), ctx.Request().Header())

				ctx.ResetContext(sctx)
				setSpanAttrs(ctx, span, ctx.Request())
				defer func() {
					tracer.End(ctx.Context(), span, reply)
				}()
			}
			return handler(ctx)
		}
	}
}

// Server ratelimiter middleware
func Client(opts ...Option) middleware.Middleware {
	options := &options{
		SpanKind:   trace.SpanKindServer,
		tracerName: _defaultName,
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	for _, o := range opts {
		o(options)
	}
	return clientByOption(options)
}

func clientByConfig(cfg *Config) middleware.Middleware {
	options := &options{
		SpanKind:   trace.SpanKindServer,
		tracerName: _defaultName,
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	return clientByOption(options)
}

// Server ratelimiter middleware
func clientByOption(options *options) middleware.Middleware {
	tracer := newTracerByOpts(options.SpanKind, options)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			if _, ok := transport.FromClientContext(ctx.Context()); ok {
				sctx, span := tracer.Start(ctx.Context(), ctx.Request().Path().FullPath(), ctx.Request().Header())

				ctx.ResetContext(sctx)
				setSpanAttrs(ctx, span, ctx.Request())
				defer func() {
					tracer.End(ctx.Context(), span, reply)
				}()
			}
			return handler(ctx)
		}
	}
}

// TraceID returns a traceid valuer.
func TraceID(ctx sctx.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		return span.TraceID().String()
	}
	return ""
}

// SpanID returns a spanid valuer.
func SpanID(ctx sctx.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.HasSpanID() {
		return span.SpanID().String()
	}
	return ""
}
