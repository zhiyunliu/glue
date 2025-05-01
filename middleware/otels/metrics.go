package otels

import (
	"github.com/zhiyunliu/glue/metrics"
)

type Metrics struct {
	RequestCounter    metrics.Int64Counter       `metric:"code_total"  `
	RequestLatency    metrics.Timer              `metric:"duration_sec"  buckets:"10,50,100,200,500,1000,2000,5000,10000" `
	RequestProcessing metrics.Int64UpDownCounter `metric:"cur_proc"  `
}
