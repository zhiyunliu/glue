package opentelemetry

import (
	"context"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/opentelemetry/metrics"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// 2. 实现可开关的 Span 处理器
type SwitchableProcessor struct {
	mu            sync.RWMutex
	enabled       bool
	realProcessor sdktrace.SpanProcessor
}

func NewSwitchableProcessor(realProcessor sdktrace.SpanProcessor) *SwitchableProcessor {
	return &SwitchableProcessor{
		enabled:       true,
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

func NewTracerProvider(serviceName string, endpoint string) *sdktrace.TracerProvider {

	res, err := resource.New(
		context.Background(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOSType(),
	)

	var opts = []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(opts...),
	)

	dynamicSampler := NewDynamicSampler(0)

	metricsObserver := metrics.NewObserver(metricsFactory, metrics.DefaultNameNormalizer)

	sampler := sdktrace.ParentBased(
		dynamicSampler, // 根Span使用动态采样
		sdktrace.WithRemoteParentSampled(sdktrace.AlwaysSample()),   // 远程父Span已采样
		sdktrace.WithRemoteParentNotSampled(sdktrace.NeverSample()), // 远程父Span未采样
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithSpanProcessor(metricsObserver),
		sdktrace.WithResource(res),
	)

}
