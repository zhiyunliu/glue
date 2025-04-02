package metrics

import (
	"sync"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/zhiyunliu/glue/metrics"
)

type MetricsEndpoints struct {
	metricsProvider  metrics.Provider
	metricsEndpoints cmap.ConcurrentMap[string, *Metrics]
	mux              sync.RWMutex
}

func NewMetricsEndpoints(metricsProvider metrics.Provider, maxNumberOfEndpoints int) *MetricsEndpoints {
	return &MetricsEndpoints{
		metricsProvider:  metricsProvider,
		metricsEndpoints: cmap.New[*Metrics](),
	}
}

func (m *MetricsEndpoints) Get(endpoint string) *Metrics {
	met, ok := m.metricsEndpoints.Get(endpoint)
	if ok {
		return met
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	//再次查询缓存，防止并发情况下重复创建metrics
	met, ok = m.metricsEndpoints.Get(endpoint)
	if ok {
		return met
	}

	met = &Metrics{}
	m.metricsProvider.Build(met)
	m.metricsEndpoints.Set(endpoint, met)
	return met
}
