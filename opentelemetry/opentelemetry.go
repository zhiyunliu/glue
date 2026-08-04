package opentelemetry

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	defaultMetricsProto = "prometheus"
	opentelemetry       = "opentelemetry"
)

// InitOtel initializes OpenTelemetry with the given service name and config.
func InitOtel(serviceName string, config config.Config) (err error) {
	telemetryConfig := config.Root().Get(opentelemetry)

	cfg := &Config{
		TraceEndpoint:   "",
		TraceSampleRate: 0,
		MetricsProvider: defaultMetricsProto,
	}
	if err := telemetryConfig.ScanTo(cfg); err != nil {
		log.Warnf("InitOtel:failed to load config: %s, use default config", err)
	}
	parpareConfig(cfg)

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
		resource.WithHost(),
	)
	if err != nil {
		err = fmt.Errorf("InitOtel:failed to create resource: %w", err)
		return err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}))
	if err = setMeterProvider(cfg.MetricsProvider); err != nil {
		return err
	}

	if err := setTracerProvider(cfg, res, telemetryConfig); err != nil {
		log.Warnf("InitOtel:%s", err)
	}
	return nil
}
