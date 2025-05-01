package prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/metrics/prometheus/collector"
	"github.com/zhiyunliu/glue/metrics"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric" // 明确重命名为 sdkmetric
)

type xResover struct {
	name string
}

func (r xResover) Name() string {
	return r.name
}
func (r *xResover) Resolve(name string, config config.Config) (metrics.Provider, error) {
	provider := &xProvider{
		config:     config,
		registerer: prometheus.DefaultRegisterer,
		gatherer:   prometheus.DefaultGatherer,
	}

	cfgObj := &prometheusConfig{
		Namespace: "server_requests",
		Job:       "microsrv",
	}
	_ = config.ScanTo(cfgObj)

	procCounter := buildProcCollector()
	if err := provider.registerer.Register(procCounter); err != nil {
		return nil, fmt.Errorf("register proc collector;err:%w", err)
	}

	exporter, err := otelprometheus.New(
		otelprometheus.WithRegisterer(provider.registerer),
		otelprometheus.WithNamespace(cfgObj.Namespace),
		otelprometheus.WithoutTargetInfo(),
		otelprometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return nil, err
	}

	// 设置 Meter Provider（使用修正后的包路径）
	meterProvider := sdkmetric.NewMeterProvider( // 使用 sdkmetric 替代 metric
		sdkmetric.WithReader(exporter),
	)

	provider.MeterProvider = meterProvider
	provider.StartPush(cfgObj, provider.gatherer)

	return provider, nil
}

func init() {

	metrics.Register(&xResover{
		name: Proto,
	})
}

// 只需要初始化一次的collector
func buildProcCollector() prometheus.Collector {
	pc, err := collector.NewProcessCPUCollector()
	if err != nil {
		err = fmt.Errorf("NewProcessCPUCollector;err:%w", err)
		panic(err)
	}
	return pc
}
