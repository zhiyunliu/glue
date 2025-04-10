package xdb

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func GetSpanFromContext(ctx context.Context) (nctx context.Context, span trace.Span) {
	tracer := otel.Tracer("XDB")
	// 创建span
	ctx, span = tracer.Start(ctx, "XDB",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	return ctx, span
}
