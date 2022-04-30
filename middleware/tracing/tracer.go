package tracing

import (
	"context"

	"github.com/zhiyunliu/gel/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Tracer is otel span tracer
type Tracer struct {
	tracer trace.Tracer
	kind   trace.SpanKind
	opt    *options
}

// NewTracer create tracer instance
func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	op := options{
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	for _, o := range opts {
		o(&op)
	}
	if op.tracerProvider != nil {
		otel.SetTracerProvider(op.tracerProvider)
	}

	return &Tracer{tracer: otel.Tracer("gel"), kind: kind, opt: &op}

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
	var err error
	err, ok := m.(error)
	if ok {

	}

	if err != nil {
		span.RecordError(err)
		if e := errors.FromError(err); e != nil {
			span.SetAttributes(attribute.Key("rpc.status_code").Int64(int64(e.Code)))
		}
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}

	if p, ok := m.(proto.Message); ok {
		if t.kind == trace.SpanKindServer {
			span.SetAttributes(attribute.Key("send_msg.size").Int(proto.Size(p)))
		} else {
			span.SetAttributes(attribute.Key("recv_msg.size").Int(proto.Size(p)))
		}
	}
	span.End()
}
