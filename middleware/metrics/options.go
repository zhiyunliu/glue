package metrics

import (
	"github.com/zhiyunliu/glue/context"
)

var CustomMetricsAttributes func(ctx context.Context) []string

func GetMetricsAttributes(ctx context.Context) []string {
	if CustomMetricsAttributes != nil {
		return CustomMetricsAttributes(ctx)
	}
	return DefaultGetMetricsAttributes(ctx)
}

func DefaultGetMetricsAttributes(ctx context.Context) []string {
	serverKind := ctx.ServerType()
	fullPath := ctx.Request().Path().FullPath()
	metricAttrs := []string{serverKind, fullPath}
	return metricAttrs

}
