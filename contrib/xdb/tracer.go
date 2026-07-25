package xdb

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/stack"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type WrapSpan interface {
	End(options ...trace.SpanEndOption)
}

func GetSpanFromContext(ctx context.Context, sting *Setting, sql, operation string, stackSkip int) (nctx context.Context, span WrapSpan) {
	meter := xdb.GetMetrics(sting.Cfg.Proto)
	meter.Incr(sting.ConnName, operation)

	tracer := otel.Tracer("XDB")
	caller := stack.Caller(stackSkip)
	// 创建span
	ctx, traceSpan := tracer.Start(ctx, fmt.Sprintf("XDB:%s", operation),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", sting.Cfg.Proto),
			attribute.String("db.conn.name", sting.ConnName),
			attribute.String("db.statement", sql),
			attribute.String("code.info", fmt.Sprintf("%x", caller)),
		),
	)
	return ctx, &WrapSpanImpl{span: traceSpan, connName: sting.ConnName, operation: operation, meter: meter}
}

type WrapSpanImpl struct {
	span      trace.Span
	meter     *xdb.Metrics
	connName  string
	operation string
}

func (w *WrapSpanImpl) End(options ...trace.SpanEndOption) {
	if w.span == nil {
		return
	}
	w.span.End(options...)
	w.meter.Decr(w.connName, w.operation)
}
