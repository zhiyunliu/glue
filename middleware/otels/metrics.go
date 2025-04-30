package otels

import (
	"github.com/zhiyunliu/glue/metrics"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	RequestCounter    metric.Int64Counter `metric:"code_total" lbls:"kind,path,code"`
	RequestLatency    metrics.Timer       `metric:"duration_sec"  lbls:"kind,path,code"`
	RequestProcessing metric.Int64Gauge   `metric:"cur_proc" lbls:"kind,path"`
}
