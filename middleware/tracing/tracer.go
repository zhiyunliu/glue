package tracing

import (
	"context"

	"github.com/zhiyunliu/glue/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is otel span tracer
type Tracer struct {
	tracer trace.Tracer
	kind   trace.SpanKind
	opt    *options
}

func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	op := &options{
		SpanKind:   kind,
		tracerName: _defaultName,
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	for i, cnt := 0, len(opts); i < cnt; i++ {
		opts[i](op)
	}
	return newTracerByOpts(kind, op)
}

// NewTracer create tracer instance
func newTracerByOpts(kind trace.SpanKind, op *options) *Tracer {

	if op.tracerProvider != nil {
		otel.SetTracerProvider(op.tracerProvider)
	}

	return &Tracer{tracer: otel.Tracer(op.tracerName), kind: kind, opt: op}

}

// Start start tracing span
func (t *Tracer) Start(ctx context.Context, path string, carrier propagation.TextMapCarrier) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer {
		ctx = t.opt.propagator.Extract(ctx, carrier)
	}
	ctx, span := t.tracer.Start(ctx,
		path,
		trace.WithSpanKind(t.kind),
	)
	if t.kind == trace.SpanKindClient {
		t.opt.propagator.Inject(ctx, carrier)
	}
	return ctx, span
}

// End finish tracing span
func (t *Tracer) End(ctx context.Context, span trace.Span, m interface{}) {
	switch val := m.(type) {
	case error:
		if e := errors.FromError(val); e != nil {
			span.SetAttributes(attribute.Key("status_code").Int64(int64(e.Code)))
		}
		span.SetStatus(codes.Error, val.Error())
	case int32:
		span.SetAttributes(attribute.Key("status_code").Int(int(val)))
	default:
		span.SetStatus(codes.Ok, "OK")
	}
	span.End()
}
