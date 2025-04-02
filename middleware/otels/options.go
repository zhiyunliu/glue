package otels

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type options struct {
	svcName string
	Kind    string
}

type config struct {
	TracerProvider oteltrace.TracerProvider
	MeterProvider  metric.MeterProvider
	Propagators    propagation.TextMapPropagator
}
