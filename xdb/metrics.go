package xdb

import (
	"context"
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	initSyncMap sync.Map
	mutex       sync.Mutex
	ScopeName   = "glue-xdb"
)

type Metrics struct {
	proto          string
	RequestCounter metrics.Int64UpDownCounter `metric:"db_cur_proc"  `
}

func GetMetrics(proto string) (meter *Metrics) {
	mutexKey := fmt.Sprintf("%s-%s", ScopeName, proto)
	meter = metrics.GetMetrics[Metrics](mutexKey, ScopeName)
	meter.proto = proto
	return meter
}

func (m *Metrics) Incr(connName, spanName string) {
	if m.RequestCounter == nil {
		return
	}
	m.RequestCounter.Add(context.Background(), 1, metric.WithAttributes(attribute.String("dbtype", m.proto), attribute.String("conn", connName), attribute.String("span", spanName)))
}

func (m *Metrics) Decr(connName, spanName string) {
	if m.RequestCounter == nil {
		return
	}
	m.RequestCounter.Add(context.Background(), -1, metric.WithAttributes(attribute.String("dbtype", m.proto), attribute.String("conn", connName), attribute.String("span", spanName)))
}
