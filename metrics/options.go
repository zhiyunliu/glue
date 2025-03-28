package metrics

type Options struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
	Buckets   []float64
}

type Option func(*Options)

func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}

}

func WithSubsystem(subsystem string) Option {
	return func(o *Options) {
		o.Subsystem = subsystem
	}
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithHelp(help string) Option {
	return func(o *Options) {
		o.Help = help
	}
}

func WithLabels(labels []string) Option {
	return func(o *Options) {
		o.Labels = labels
	}
}

func WithBuckets(buckets []float64) Option {
	return func(o *Options) {
		o.Buckets = buckets
	}
}
