package otels

import (
	"github.com/zhiyunliu/glue/context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type options struct {
	svcName string
}

type config struct {
	TracerProvider              oteltrace.TracerProvider
	Propagators                 propagation.TextMapPropagator
	Filters                     []Filter
	MeterProvider               metric.MeterProvider
	AdditionalAttributeCallback MetricAttributeFn
}

type Filter func(context.Context) bool
type MetricAttributeFn func(req context.Request) []attribute.KeyValue
