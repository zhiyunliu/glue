package dlocker

type Options struct {
	Data string
}

type Option func(opts *Options)

func WithData(data string) Option {
	return func(opts *Options) {
		opts.Data = data
	}
}
