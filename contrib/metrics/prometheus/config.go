package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type prometheusConfig struct {
	Counter   *counterOpts   `json:"counter"`
	Histogram *histogramOpts `json:"histogram"`
	Gauge     *gaugeOpts     `json:"gauge"`
}

type counterOpts struct {
	Namespace string   `json:"namespace"`
	Subsystem string   `json:"subsystem"`
	Name      string   `json:"name"`
	Help      string   `json:"help"`
	Labels    []string `json:"labels"`
}

type histogramOpts struct {
	Namespace string    `json:"namespace"`
	Subsystem string    `json:"subsystem"`
	Name      string    `json:"name"`
	Help      string    `json:"help"`
	Buckets   []float64 `json:"buckets"`
	Labels    []string  `json:"labels"`
}

type gaugeOpts struct {
	Namespace string   `json:"namespace"`
	Subsystem string   `json:"subsystem"`
	Name      string   `json:"name"`
	Help      string   `json:"help"`
	Labels    []string `json:"labels"`
}

func (c *prometheusConfig) GetCounter() (opts prometheus.CounterOpts, labels []string) {
	ctr := c.Counter
	return prometheus.CounterOpts{
		Namespace: ctr.Namespace,
		Subsystem: ctr.Subsystem,
		Name:      ctr.Name,
		Help:      ctr.Help,
	}, ctr.Labels
}

func (c *prometheusConfig) GetHistogram() (opts prometheus.HistogramOpts, labels []string) {
	ctr := c.Histogram
	return prometheus.HistogramOpts{
		Namespace: ctr.Namespace,
		Subsystem: ctr.Subsystem,
		Name:      ctr.Name,
		Help:      ctr.Help,
		Buckets:   ctr.Buckets,
	}, ctr.Labels
}

func (c *prometheusConfig) GetGauge() (opts prometheus.GaugeOpts, labels []string) {
	ctr := c.Gauge
	return prometheus.GaugeOpts{
		Namespace: ctr.Namespace,
		Subsystem: ctr.Subsystem,
		Name:      ctr.Name,
		Help:      ctr.Help,
	}, ctr.Labels
}
