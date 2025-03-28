package prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/config"
	collector "github.com/zhiyunliu/glue/contrib/metrics/prometheus/collector"
	"github.com/zhiyunliu/glue/metrics"
)

type xResover struct {
	factory metrics.Factory
}

func (r xResover) Name() string {
	return Proto
}
func (r *xResover) Resolve(name string, config config.Config) (metrics.Provider, error) {

	provider := &xProvider{
		factory: r.factory,
		config:  config,
	}

	procCounter := buildProcCollector()
	prometheus.MustRegister(procCounter)

	return provider, nil
}

// 只需要初始化一次的collector
func buildProcCollector() prometheus.Collector {
	pc, err := collector.NewProcessCollector()
	if err != nil {
		err = fmt.Errorf("NewProcessCollector;err:%w", err)
		panic(err)
	}
	return pc
}

func init() {
	metrics.Register(&xResover{
		factory: NewFactory(
			WithNameSpace("server"),
			WithSubsystem("requests"),
			WithRegisterer(prometheus.DefaultRegisterer),
			WithDefaultBuckets(0.05, 0.1, 0.5, 1, 1.5, 2, 2.5, 3, 4, 5),
		),
	})
}
