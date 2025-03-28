package semconv

import (
	"github.com/zhiyunliu/glue/context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func RequestTraceAttrs(server string, req context.Request) []attribute.KeyValue {
	count := 3 // ServerAddress, Method, Scheme

	clientIP := req.GetClientIP()

	attrs := make([]attribute.KeyValue, 0, count)
	attrs = append(attrs,
		semconv.HTTPRequestMethodKey.String(req.GetMethod()),
	)

	if useragent := req.GetHeader("User-Agent"); useragent != "" {
		attrs = append(attrs, semconv.UserAgentName(useragent))
	}

	if clientIP != "" {
		attrs = append(attrs, semconv.ClientAddress(clientIP))
	}

	if req.Path().GetURL() != nil && req.Path().GetURL().Path != "" {
		attrs = append(attrs, semconv.URLPath(req.Path().GetURL().Path))
	}

	return attrs
}
