package prometheus

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"golang.org/x/sync/errgroup"
)

type prometheusConfig struct {
	Gateway   *gateway       `json:"gateway"`
	Counter   *counterOpts   `json:"counter"`
	Histogram *histogramOpts `json:"histogram"`
	Gauge     *gaugeOpts     `json:"gauge"`
}

type gateway struct {
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
}

type normalOpts struct {
	Namespace string   `json:"namespace"`
	Subsystem string   `json:"subsystem"`
	Name      string   `json:"name"`
	Help      string   `json:"help"`
	Labels    []string `json:"labels"`
}

type counterOpts struct {
	normalOpts
}

type histogramOpts struct {
	normalOpts
	Buckets []float64 `json:"buckets"`
}

type gaugeOpts struct {
	normalOpts
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

func (c *prometheusConfig) StartPush(collectors ...prometheus.Collector) {
	//没有设置pushgateway，则不启动push功能
	if c.Gateway == nil {
		return
	}
	if global.HasApi {
		return
	}

	if c.Gateway.Addr == "" {
		c.Gateway.Addr = os.Getenv("PROMETHEUS_PUSH_GATEWAY_ADDR")
	}

	if c.Gateway.Addr == "" {
		log.Warnf("Pushgateway Addr is not set, Prometheus push is disabled")
		return
	}

	group := errgroup.Group{}
	group.Go(func() error {
		cfg := c.Gateway

		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		pusher := push.New(cfg.Addr, "microsrv").
			Grouping("instance", global.LocalIp).
			Grouping("srv", global.AppName)

		for _, collector := range collectors {
			pusher.Collector(collector)
		}

		for range ticker.C {
			c.execPush(pusher)
		}
		return nil
	})
}

func (c *prometheusConfig) execPush(pusher *push.Pusher) {
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
