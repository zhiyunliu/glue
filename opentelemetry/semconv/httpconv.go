package semconv

import (
	"github.com/zhiyunliu/glue/context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func RequestTraceAttrs(ctx context.Context) []attribute.KeyValue {
	count := 7 // ServerAddress, Method, Scheme
	req := ctx.Request()
	clientIP := req.GetClientIP()

	attrs := make([]attribute.KeyValue, 0, count)
	attrs = append(attrs,
		attribute.String("content.type", req.ContentType()),
		attribute.String("server.type", ctx.ServerType()),
		semconv.HTTPRequestMethodKey.String(req.GetMethod()),
		XRequestID(req.RequestID()),
	)
	query := req.Query()
	if qval := query.GetValues(); len(qval) > 0 {
		attrs = append(attrs, semconv.URLQuery(qval.Encode()))
	}

	if useragent := req.GetHeader("User-Agent"); useragent != "" {
		attrs = append(attrs, semconv.UserAgentName(useragent))
	}

	if clientIP != "" {
		attrs = append(attrs, semconv.ClientAddress(clientIP))
	}

	return attrs
}
