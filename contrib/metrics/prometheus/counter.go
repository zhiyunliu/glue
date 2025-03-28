package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/metrics"
)

var _ metrics.Counter = (*counter)(nil)

type counter struct {
	cv *prometheus.CounterVec
}

// NewCounter new a prometheus counter and returns Counter.
func NewCounter(cv *prometheus.CounterVec) metrics.Counter {
	return &counter{
		cv: cv,
	}
}

func (c *counter) Inc(lbvs ...string) {
	c.cv.WithLabelValues(lbvs...).Inc()
}

func (c *counter) Add(delta float64, lbvs ...string) {
	c.cv.WithLabelValues(lbvs...).Add(delta)
}
