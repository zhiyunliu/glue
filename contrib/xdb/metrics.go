package xdb

import (
	"context"
	"sync"

	"github.com/zhiyunliu/glue/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	initSyncMap sync.Map
	mutex       sync.Mutex
	ScopeName   = "glue-xdb"
)

type Metrics struct {
	Proto          string
	RequestCounter metrics.Int64UpDownCounter `metric:"db_cur_proc"  `
}

func GetMetrics(proto string) (meter *Metrics) {
	tmp, ok := initSyncMap.Load(proto)
	if ok {
		return tmp.(*Metrics)
	}
	mutex.Lock()
	defer mutex.Unlock()

	tmp, ok = initSyncMap.Load(proto)
	if ok {
		return tmp.(*Metrics)
	}

	meter = &Metrics{
		Proto: proto,
	}

	factory := metrics.NewFactory(otel.GetMeterProvider(), ScopeName)
	metrics.Init(meter, factory)
	initSyncMap.Store(proto, meter)
	return meter
}

func (m *Metrics) Incr(connName, spanName string) {
	if m.RequestCounter == nil {
		return
	}
	m.RequestCounter.Add(context.Background(), 1, metric.WithAttributes(attribute.String("dbtype", m.Proto), attribute.String("conn", connName), attribute.String("span", spanName)))
}

func (m *Metrics) Decr(connName, spanName string) {
	if m.RequestCounter == nil {
		return
	}
	m.RequestCounter.Add(context.Background(), -1, metric.WithAttributes(attribute.String("dbtype", m.Proto), attribute.String("conn", connName), attribute.String("span", spanName)))
}
