package xdb

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/stack"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func GetSpanFromContext(ctx context.Context, sting *Setting, sql, operation string, stackSkip int) (nctx context.Context, span trace.Span) {
	tracer := otel.Tracer("XDB")
	caller := stack.Caller(stackSkip)
	// 创建span
	ctx, span = tracer.Start(ctx, fmt.Sprintf("XDB:%s", operation),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", sting.Cfg.Proto),
			attribute.String("db.conn.name", sting.ConnName),
			attribute.String("db.statement", sql),
			attribute.String("code.info", fmt.Sprintf("%x", caller)),
		),
	)
	return ctx, span
}
