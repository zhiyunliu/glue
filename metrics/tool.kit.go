package metrics

import (
	"sync"

	"go.opentelemetry.io/otel"
)

var (
	initSyncMap sync.Map
	mutex       sync.Mutex
)

// InitMetrics 初始化指标
// @param mutexKey 指标缓存key
// @return 指标实例
func GetMetrics[T any](mutexKey, metricName string) (meter *T) {
	if mutexKey == "" {
		return
	}

	tmp, ok := initSyncMap.Load(mutexKey)
	if ok {
		return tmp.(*T)
	}
	mutex.Lock()
	defer mutex.Unlock()

	tmp, ok = initSyncMap.Load(mutexKey)
	if ok {
		return tmp.(*T)
	}

	meter = new(T)

	factory := NewFactory(otel.GetMeterProvider(), metricName)
	_ = Init(meter, factory)
	initSyncMap.Store(mutexKey, meter)
	return meter
}
