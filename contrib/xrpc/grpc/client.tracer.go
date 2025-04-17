package grpc

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/global"
	gsemconv "github.com/zhiyunliu/glue/opentelemetry/semconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/trace"
)

func getTracerAttributes(req *grpcproto.Request) []attribute.KeyValue {
	attris := make([]attribute.KeyValue, 0, 5)
	// 设置请求头属性
	attris = append(attris,
		attribute.String("rpc.system", Proto),
		attribute.String("rpc.service", fmt.Sprintf("%s %s://%s%s", req.Method, Proto, global.AppName, req.Service)),
	)

	if xrequestId := req.Header[constants.HeaderRequestId]; xrequestId != "" {
		attris = append(attris, gsemconv.XRequestID(xrequestId)) // 添加X-Request-ID到span的属性
	}
	return attris
}

func GetNormalSpanFromContext(ctx context.Context, req *grpcproto.Request) (nctx context.Context, span trace.Span) {
	attris := getTracerAttributes(req)

	tracer := otel.Tracer("grpc-normal-client")
	// 创建span
	ctx, span = tracer.Start(ctx,
		"GRPC Normal Client "+req.Service,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attris...),
	)
	// 注入跟踪信息到请求头
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, engine.Header(req.Header))

	return ctx, span
}

func GetStreamSpanFromContext(ctx context.Context, req *grpcproto.Request) (nctx context.Context, span trace.Span) {
	attris := getTracerAttributes(req)

	tracer := otel.Tracer("grpc-stream-client")
	// 创建span
	ctx, span = tracer.Start(ctx,
		"GRPC Stream Client "+req.Service,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attris...),
	)
	// 注入跟踪信息到请求头
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.MapCarrier(req.Header))

	return ctx, span
}
