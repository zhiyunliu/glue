package semconv

import (
	"fmt"

	"github.com/zhiyunliu/glue/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func Status(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

type ServerMetricData struct {
	ServerName   string
	ResponseSize int64

	MetricData
	MetricAttributes
}

type MetricAttributes struct {
	Req                  context.Request
	StatusCode           int
	AdditionalAttributes []attribute.KeyValue
}

type MetricData struct {
	RequestSize int64

	// The request duration, in milliseconds
	ElapsedTime float64
}
