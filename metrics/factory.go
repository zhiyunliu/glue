package metrics

import "github.com/zhiyunliu/glue/config"

type Factory interface {
	Counter(metric config.Config, opts *Options) Counter
	Timer(metric config.Config, opts *Options) Timer
	Gauge(metric config.Config, opts *Options) Gauge
	Histogram(metric config.Config, opts *Options) Histogram
}

var noopFactory Factory = xnoopFactory{}

type xnoopFactory struct{}

func (xnoopFactory) Counter(config.Config, *Options) Counter {
	return NoopCounter
}

func (xnoopFactory) Timer(config.Config, *Options) Timer {
	return NoopTimer
}

func (xnoopFactory) Gauge(config.Config, *Options) Gauge {
	return NoopGauge
}

func (xnoopFactory) Histogram(config.Config, *Options) Histogram {
	return NoopHistogram
}
