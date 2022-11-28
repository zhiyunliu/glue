package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/metrics"
)

const (
	Proto = "prometheus"
)

type xProvider struct {
	counter  metrics.Counter
	observer metrics.Observer
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

func (p xProvider) GetImpl() interface{} {
	return Proto

}

type xResover struct {
}

func (r xResover) Name() string {
	return Proto
}
func (r xResover) Resolve(name string, config config.Config) (metrics.Provider, error) {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "The total number of processed requests",
	}, []string{"kind", "path", "code", "reason"})

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "duration_sec",
		Help:      "server requests duration(sec).",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
	}, []string{"kind", "path"})

	prometheus.MustRegister(counter)
	prometheus.MustRegister(histogram)

	return &xProvider{
		counter:  NewCounter(counter),
		observer: NewHistogram(histogram),
	}, nil
}

func init() {
	metrics.Register(&xResover{})

}
