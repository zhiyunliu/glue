package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/zhiyunliu/glue/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func setTracerProvider(cfg *Config, res *resource.Resource, telemetryConfig config.Config) error {
	var opts = []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(opts...),
	)
	if err != nil {
		err = fmt.Errorf("setTracerProvider.failed to create exporter: %w", err)
		return err
	}

	dynamicSampler := newDynamicSampler(cfg.SamplerRate, telemetryConfig)
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
