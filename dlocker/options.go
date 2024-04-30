package dlocker

type Options struct {
	Data string
	AutoRenewal bool 
}

type Option func(opts *Options)

//设置数据
func WithData(data string) Option {
	return func(opts *Options) {
		opts.Data = data
	}
}

//自动续期
func WithAutoRenewal() Option {
	return func(opts *Options) {
		opts.AutoRenewal = true
	}
}
