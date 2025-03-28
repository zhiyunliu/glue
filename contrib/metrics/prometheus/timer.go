package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/metrics"
)

var _ metrics.Timer = (*timer)(nil)

type timer struct {
	hv *prometheus.HistogramVec
}

func (h *timer) Record(value time.Duration, lbvs ...string) {
	h.hv.WithLabelValues(lbvs...).Observe(float64(value.Milliseconds()))
}
