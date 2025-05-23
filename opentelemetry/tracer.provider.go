package opentelemetry

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func setTracerProvider(cfg *Config, res *resource.Resource, telemetryConfig config.Config) error {

	exporter, err := buildTraceExporter(cfg)
	if err != nil {
		err = fmt.Errorf("setTracerProvider.failed to create exporter: %w", err)
		return err
	}

	dynamicSampler := newDynamicSampler(cfg.TraceSampleRate, telemetryConfig)
	err = dynamicSampler.Watch()
	if err != nil {
		err = fmt.Errorf("setTracerProvider.failed to watch sampler rate: %w", err)
		return err
	}

	sampler := sdktrace.ParentBased(
		dynamicSampler, // 根Span使用动态采样
		sdktrace.WithRemoteParentSampled(sdktrace.AlwaysSample()),   // 远程父Span已采样
		sdktrace.WithRemoteParentNotSampled(sdktrace.NeverSample()), // 远程父Span未采样
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return nil

}

func buildTraceExporter(cfg *Config) (exporter sdktrace.SpanExporter, err error) {
	if cfg.TraceEndpoint == "" {
		err = fmt.Errorf("buildTraceExporter: endpoint is empty")
		return
	}

	urlObj, err := url.Parse(cfg.TraceEndpoint)
	if err != nil {
		err = fmt.Errorf("buildTraceExporter: failed to parse endpoint: %w", err)
		return
	}
	factory, ok := exporterMap[urlObj.Scheme]
	if !ok {
		return nil, fmt.Errorf("buildTraceExporter: unsupported scheme: %s", urlObj.Scheme)
	}
	return factory(cfg)
}

type SpanExporterFactory func(cfg *Config) (sdktrace.SpanExporter, error)

var (
	exporterMap = make(map[string]SpanExporterFactory)
)

func init() {

	exporterMap["http"] = func(cfg *Config) (sdktrace.SpanExporter, error) {
		urlObj, _ := url.Parse(cfg.TraceEndpoint)
		var opts = []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(urlObj.Host),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exporter, err := otlptrace.New(
			context.Background(),
			otlptracehttp.NewClient(opts...),
		)
		return exporter, err
	}
	exporterMap["grpc"] = func(cfg *Config) (sdktrace.SpanExporter, error) {
		urlObj, _ := url.Parse(cfg.TraceEndpoint)
		var opts = []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(urlObj.Host),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		exporter, err := otlptrace.New(
			context.Background(),
			otlptracegrpc.NewClient(opts...),
		)
		return exporter, err

	}

}
