package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// Factory provides factory methods for creating metric instruments
type Factory struct {
	meter metric.Meter
}

// NewFactory creates a new Factory from a MeterProvider
func NewFactory(meterProvider metric.MeterProvider, meterName string) *Factory {
	return &Factory{
		meter: meterProvider.Meter(meterName),
	}
}

// CreateCounter creates a new Counter instrument
func (f *Factory) CreateIntCounter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return f.meter.Int64Counter(name, opts...)
}

// CreateFloatCounter creates a new Float64 Counter instrument
func (f *Factory) CreateFloatCounter(name string, opts ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return f.meter.Float64Counter(name, opts...)
}

// CreateGauge creates a new Gauge instrument (using ObservableGauge)
func (f *Factory) CreateIntGauge(name string, opts ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	return f.meter.Int64Gauge(name, opts...)
}

// CreateIntGauge creates a new integer Gauge instrument
func (f *Factory) CreateFloatGauge(name string, opts ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	return f.meter.Float64Gauge(name, opts...)
}

// CreateHistogram creates a new Histogram instrument
func (f *Factory) CreateFloatHistogram(name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return f.meter.Float64Histogram(name, opts...)
}

// CreateIntHistogram creates a new integer Histogram instrument
func (f *Factory) CreateIntHistogram(name string, opts ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return f.meter.Int64Histogram(name, opts...)
}

// CreateTimer creates a new Timer (wrapped Histogram for measuring durations)
func (f *Factory) CreateTimer(name string, opts ...metric.Float64HistogramOption) (Timer, error) {
	histogram, err := f.CreateFloatHistogram(name, opts...)
	if err != nil {
		return nil, err
	}

	return &xTimer{histogram: histogram}, nil

}

type Timer interface {
	Record(ctx context.Context, start time.Time, opts ...metric.RecordOption)
}

// Timer is a convenience wrapper for measuring durations
type xTimer struct {
	histogram metric.Float64Histogram
}

// Record records the duration since the specified start time
func (t *xTimer) Record(ctx context.Context, start time.Time, opts ...metric.RecordOption) {
	duration := float64(time.Since(start).Microseconds()) / 1000.0 // convert to milliseconds
	t.histogram.Record(context.Background(), duration, opts...)
}
