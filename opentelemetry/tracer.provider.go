package opentelemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	defaultMetricsProto = "prometheus"
	opentelemetry       = "opentelemetry"
)

// 2. 实现可开关的 Span 处理器
type SwitchableProcessor struct {
	mu            sync.RWMutex
	enabled       bool
	realProcessor sdktrace.SpanProcessor
}

func NewSwitchableProcessor(realProcessor sdktrace.SpanProcessor) *SwitchableProcessor {
	return &SwitchableProcessor{
		enabled:       false,
		realProcessor: realProcessor,
	}
}

func (p *SwitchableProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.enabled {
		p.realProcessor.OnStart(ctx, s)
	}
}

func (p *SwitchableProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.enabled {
		p.realProcessor.OnEnd(s)
	}
}

func (p *SwitchableProcessor) Shutdown(ctx context.Context) error {
	return p.realProcessor.Shutdown(ctx)
}

func (p *SwitchableProcessor) ForceFlush(ctx context.Context) error {
	return p.realProcessor.ForceFlush(ctx)
}

func (p *SwitchableProcessor) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
}

// NewTracerProvider creates a new TracerProvider with the given configuration.
func NewTracerProvider(serviceName string, config config.Config) (*sdktrace.TracerProvider, error) {
	telemetryConfig := config.Root().Get(opentelemetry)

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		err = fmt.Errorf("failed to create resource: %w", err)
		return nil, err
	}
	cfg := &Config{}
	err = telemetryConfig.ScanTo(cfg)
	if err != nil {
		err = fmt.Errorf("failed to load config: %w", err)
		return nil, err
	}
	var opts = []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(opts...),
	)
	if err != nil {
		err = fmt.Errorf("failed to create exporter: %w", err)
		return nil, err
	}

	dynamicSampler := NewDynamicSampler(serviceName, cfg.SamplerRate, telemetryConfig)
	err = dynamicSampler.Watch()
	if err != nil {
		err = fmt.Errorf("failed to watch sampler rate: %w", err)
		return nil, err
	}

	sampler := sdktrace.ParentBased(
		dynamicSampler, // 根Span使用动态采样
		sdktrace.WithRemoteParentSampled(sdktrace.AlwaysSample()),   // 远程父Span已采样
		sdktrace.WithRemoteParentNotSampled(sdktrace.NeverSample()), // 远程父Span未采样
	)

	metricsProto := cfg.MetricsProto
	if metricsProto == "" {
		metricsProto = defaultMetricsProto
	}

 
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
 		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}))

	return tp, nil

}
