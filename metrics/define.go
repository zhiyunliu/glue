package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
)

type Int64Counter = metric.Int64Counter
type Int64UpDownCounter = metric.Int64UpDownCounter
type Float64Counter = metric.Float64Counter
type Int64Gauge = metric.Int64Gauge
type Float64Gauge = metric.Float64Gauge
type Float64Histogram = metric.Float64Histogram
type Int64Histogram = metric.Int64Histogram
type Timer interface {
	Record(ctx context.Context, start time.Time, opts ...metric.RecordOption)
}
