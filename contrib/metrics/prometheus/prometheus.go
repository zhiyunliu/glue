package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/metrics"
)

const (
	Proto = "prometheus"
)

var _ metrics.Provider = &xProvider{}

type xProvider struct {
	counter  metrics.Counter
	observer metrics.Observer
	gauge    metrics.Gauge
}

func (p xProvider) Name() string {
	return Proto
}

func (p xProvider) Counter() metrics.Counter {
	return p.counter

}

func (p xProvider) Observer() metrics.Observer {
	return p.observer
}
func (p xProvider) Gauge() metrics.Gauge {
	return p.gauge
}

func (p xProvider) GetImpl() interface{} {
	return Proto

}

type xResover struct {
}

func (r xResover) Name() string {
	return Proto
}
func (r xResover) Resolve(name string, config config.Config) (metrics.Provider, error) {
	configOpts := prometheusConfig{
		Counter: &counterOpts{
			Namespace: "server",
			Subsystem: "requests",
			Name:      "code_total",
			Help:      "The total number of processed requests",
			Labels:    []string{"kind", "path", "code", "reason"},
		},
		Histogram: &histogramOpts{
			Namespace: "server",
			Subsystem: "requests",
			Name:      "duration_sec",
			Help:      "server requests duration(sec).",
			Buckets:   []float64{0.05, 0.1, 0.5, 1, 1.5, 2, 2.5, 3, 4, 5},
			Labels:    []string{"kind", "path"},
		},
		Gauge: &gaugeOpts{
			Namespace: "server",
			Subsystem: "requests",
			Name:      "cur_proc",
			Help:      "server current processing.",
			Labels:    []string{"kind", "path"},
		},
	}
	config.Scan(&configOpts)

	counter := prometheus.NewCounterVec(configOpts.GetCounter())
	histogram := prometheus.NewHistogramVec(configOpts.GetHistogram())
	gauge := prometheus.NewGaugeVec(configOpts.GetGauge())

	prometheus.MustRegister(counter)
	prometheus.MustRegister(histogram)
	prometheus.MustRegister(gauge)

	return &xProvider{
		counter:  NewCounter(counter),
		observer: NewHistogram(histogram),
		gauge:    NewGauge(gauge),
	}, nil
}

func init() {
	metrics.Register(&xResover{})

}
