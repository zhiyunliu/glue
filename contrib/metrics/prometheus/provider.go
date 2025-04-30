package prometheus

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/metrics"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/sync/errgroup"
)

const (
	Proto = "prometheus"
)

var (
	_ metrics.Provider = &xProvider{}
)

type xProvider struct {
	config config.Config
	metric.MeterProvider
}

func (p xProvider) Name() string {
	return Proto
}

func (p *xProvider) GetImpl() any {
	return p.MeterProvider
}

func (p *xProvider) StartPush(config *prometheusConfig, collectors ...prometheus.Collector) {
	if config.Gateway == nil {
		config.Gateway = &gateway{
			Addr:     os.Getenv("PROMETHEUS_PUSH_GATEWAY_ADDR"),
			Interval: 15,
		}
	}
	if global.HasApi {
		return
	}

	if config.Gateway.Addr == "" {
		log.Warnf("Pushgateway Addr is not set, Prometheus push is disabled")
		return
	}

	group := errgroup.Group{}
	group.Go(func() error {
		cfg := config.Gateway

		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		pusher := push.New(cfg.Addr, config.Job).
			Grouping("instance", global.LocalIp).
			Grouping("srv", global.AppName)

		for _, collector := range collectors {
			pusher.Collector(collector)
		}

		for range ticker.C {
			p.execPush(pusher)
		}
		return nil
	})
}

func (p *xProvider) execPush(pusher *push.Pusher) {
	defer func() {
		if obj := recover(); obj != nil {
			log.Panicf("Push metrics to Pushgateway: %v", obj)
		}
	}()
	// 将指标推送到 Pushgateway
	if err := pusher.Push(); err != nil {
		log.Warnf("Could not push metrics to Pushgateway: %v", err)
	}
}
