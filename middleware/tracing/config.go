package tracing

import "go.opentelemetry.io/otel/trace"

//	{"provider":"skywalking","propagator":"propagator"}
type Config struct {
	SpanKind   trace.SpanKind `json:"span_kind" yaml:"span_kind"`
	Provider   string         `json:"provider" yaml:"provider"`
	Propagator string         `json:"propagator" yaml:"propagator"`
}
