package otels

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/metrics"
	"go.opentelemetry.io/otel/attribute"
)

type Metrics struct {
	RequestCounter    metrics.Int64Counter       `metric:"code_total"  `
	RequestLatency    metrics.Timer              `metric:"duration_sec"  buckets:"10,50,100,200,500,1000,2000,5000,10000" `
	RequestProcessing metrics.Int64UpDownCounter `metric:"cur_proc"  `
}

var CustomMetricsAttributes func(ctx context.Context) []attribute.KeyValue

func GetMetricsAttributes(ctx context.Context) []attribute.KeyValue {
	if CustomMetricsAttributes != nil {
		return CustomMetricsAttributes(ctx)
	}

	return DefaultGetMetricsAttributes(ctx)
}

func DefaultGetMetricsAttributes(ctx context.Context) []attribute.KeyValue {
	serverKind := ctx.ServerType()
	fullPath := ctx.Request().Path().FullPath()
	metricAttrs := []attribute.KeyValue{
		attribute.String("kind", serverKind),
		attribute.String("path", fullPath),
	}
	return metricAttrs

}
