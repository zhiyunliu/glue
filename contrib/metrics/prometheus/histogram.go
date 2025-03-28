package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/metrics"
)

var _ metrics.Histogram = (*histogram)(nil)

type histogram struct {
	hv *prometheus.HistogramVec
}

// NewHistogram new a prometheus histogram and returns Histogram.

func (h *histogram) Record(value float64, lbvs ...string) {
	h.hv.WithLabelValues(lbvs...).Observe(value)
}
