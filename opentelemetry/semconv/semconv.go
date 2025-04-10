package semconv

import "go.opentelemetry.io/otel/attribute"

const (
	XRequestIDKey = attribute.Key("x.request.id")
	MethodKey     = attribute.Key("http.method")
)

func XRequestID(val string) attribute.KeyValue {
	return XRequestIDKey.String(val)
}

func Method(val string) attribute.KeyValue {
	return MethodKey.String(val)
}
