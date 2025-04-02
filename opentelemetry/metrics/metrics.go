package metrics

import (
	"github.com/zhiyunliu/glue/metrics"
)

type Metrics struct {
	RequestCounter    metrics.Counter `metric:"code_total" lbls:"kind,path,code"`
	RequestLatency    metrics.Timer   `metric:"duration_sec"  lbls:"kind,path,code"`
	RequestProcessing metrics.Gauge   `metric:"cur_proc" lbls:"kind,path"`
}
