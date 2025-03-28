package prometheus

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/metrics"
)

var _ metrics.Factory = (*Factory)(nil)

type options struct {
	registerer     prometheus.Registerer
	defaultBuckets []float64
	separator      string
	namespace      string
	subsystem      string
}

type Option func(*options)

func WithNameSpace(space string) Option {
	return func(opts *options) {
		opts.namespace = space
	}
}

func WithSubsystem(subsys string) Option {
	return func(opts *options) {
		opts.subsystem = subsys
	}
}

func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(opts *options) {
		opts.registerer = registerer
	}
}

func WithDefaultBuckets(buckets ...float64) Option {
	return func(opts *options) {
		opts.defaultBuckets = buckets
	}
}

type Factory struct {
	scope      string
	cache      *metricCache
	normalizer *strings.Replacer
	options    *options
}

func NewFactory(opts ...Option) metrics.Factory {
	options := new(options)
	for _, o := range opts {
		o(options)
	}
	if options.registerer == nil {
		options.registerer = prometheus.DefaultRegisterer
	}

	return &Factory{
		cache:      newCache(options.registerer),
		normalizer: strings.NewReplacer(".", "_", "-", "_"),
		options:    options,
	}
}

func (f *Factory) Counter(cfg config.Config, opts *metrics.Options) metrics.Counter {

	copts := prometheus.CounterOpts{
		Namespace: f.getNamespace(opts.Namespace),
		Subsystem: f.getNamespace(opts.Subsystem),
		Name:      opts.Name,
		Help:      f.getHelp(opts),
	}
	cv := f.cache.getOrCreateCounter(copts, opts.Labels)
	return &counter{
		cv: cv,
	}
}

func (f *Factory) Timer(cfg config.Config, opts *metrics.Options) metrics.Timer {
	hopts := prometheus.HistogramOpts{
		Namespace: f.getNamespace(opts.Namespace),
		Subsystem: f.getNamespace(opts.Subsystem),
		Name:      opts.Name,
		Help:      f.getHelp(opts),
		Buckets:   f.getBuckets(opts.Buckets),
	}
	hv := f.cache.getOrCreateHistogram(hopts, opts.Labels)
	return &timer{
		hv: hv,
	}
}

func (f *Factory) Gauge(cfg config.Config, opts *metrics.Options) metrics.Gauge {

	gopts := prometheus.GaugeOpts{
		Namespace: f.getNamespace(opts.Namespace),
		Subsystem: f.getNamespace(opts.Subsystem),
		Name:      opts.Name,
		Help:      f.getHelp(opts),
	}
	gv := f.cache.getOrCreateGauge(gopts, opts.Labels)
	return &gauge{
		gv: gv,
	}
}

func (f *Factory) Histogram(cfg config.Config, opts *metrics.Options) metrics.Histogram {

	hopts := prometheus.HistogramOpts{
		Namespace: f.getNamespace(opts.Namespace),
		Subsystem: f.getNamespace(opts.Subsystem),
		Name:      opts.Name,
		Help:      f.getHelp(opts),
		Buckets:   f.getBuckets(opts.Buckets),
	}
	hv := f.cache.getOrCreateHistogram(hopts, opts.Labels)
	return &histogram{
		hv: hv,
	}
}

func (f *Factory) getNamespace(namespace string) string {
	if len(namespace) > 0 {
		return namespace
	}
	return f.options.namespace
}

func (f *Factory) getSubSystem(subsystem string) string {
	if len(subsystem) > 0 {
		return subsystem
	}
	return f.options.subsystem
}
func (f *Factory) getHelp(opts *metrics.Options) string {
	help := strings.TrimSpace(opts.Help)
	if len(help) == 0 {
		help = opts.Name
	}
	return help
}
func (f *Factory) getBuckets(buckets []float64) []float64 {
	if len(buckets) > 0 {
		return buckets
	}
	return f.options.defaultBuckets
}
