package metrics

type Options struct {
	Name    string
	Help    string
	Buckets []float64
}

type Option func(*Options)

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

func WithBuckets(buckets []float64) Option {
	return func(o *Options) {
		o.Buckets = buckets
	}
}
