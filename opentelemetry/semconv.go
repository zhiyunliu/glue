package opentelemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	XRequestIDKey = attribute.Key("x.request.id")
)

func XRequestID(val string) attribute.KeyValue {
	return XRequestIDKey.String(val)
}

func Status(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}
