package prometheus

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type metricCache struct {
	registerer prometheus.Registerer
	lock       sync.Mutex
	cVecs      map[string]*prometheus.CounterVec
	gVecs      map[string]*prometheus.GaugeVec
	hVecs      map[string]*prometheus.HistogramVec
}

func newCache(registerer prometheus.Registerer) *metricCache {
	return &metricCache{
		registerer: registerer,
		cVecs:      make(map[string]*prometheus.CounterVec),
		gVecs:      make(map[string]*prometheus.GaugeVec),
		hVecs:      make(map[string]*prometheus.HistogramVec),
	}
}

func (c *metricCache) getOrCreateCounter(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	c.lock.Lock()
	defer c.lock.Unlock()

	cacheKey := c.getCacheKey(opts.Name, labelNames)
	cv, cvExists := c.cVecs[cacheKey]
	if !cvExists {
		cv = prometheus.NewCounterVec(opts, labelNames)
		c.registerer.MustRegister(cv)
		c.cVecs[cacheKey] = cv
	}
	return cv
}

func (c *metricCache) getOrCreateGauge(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	c.lock.Lock()
	defer c.lock.Unlock()

	cacheKey := c.getCacheKey(opts.Name, labelNames)
	gv, gvExists := c.gVecs[cacheKey]
	if !gvExists {
		gv = prometheus.NewGaugeVec(opts, labelNames)
		c.registerer.MustRegister(gv)
		c.gVecs[cacheKey] = gv
	}
	return gv
}

func (c *metricCache) getOrCreateHistogram(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	c.lock.Lock()
	defer c.lock.Unlock()

	cacheKey := c.getCacheKey(opts.Name, labelNames)
	hv, hvExists := c.hVecs[cacheKey]
	if !hvExists {
		hv = prometheus.NewHistogramVec(opts, labelNames)
		c.registerer.MustRegister(hv)
		c.hVecs[cacheKey] = hv
	}
	return hv
}

func (*metricCache) getCacheKey(name string, labels []string) string {
	return strings.Join(append([]string{name}, labels...), ":")
}
