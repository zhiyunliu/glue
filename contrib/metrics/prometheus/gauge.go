package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/metrics"
)

var _ metrics.Gauge = (*gauge)(nil)

type gauge struct {
	gv *prometheus.GaugeVec
}

// NewGauge new a prometheus gauge and returns Gauge.
func NewGauge(gv *prometheus.GaugeVec) metrics.Gauge {
	return &gauge{
		gv: gv,
	}
}

func (g *gauge) Set(value float64, lbvs ...string) {
	g.gv.WithLabelValues(lbvs...).Set(value)
}

func (g *gauge) Add(delta float64, lbvs ...string) {
	g.gv.WithLabelValues(lbvs...).Add(delta)
}

func (g *gauge) Sub(delta float64, lbvs ...string) {
	g.gv.WithLabelValues(lbvs...).Sub(delta)
}
