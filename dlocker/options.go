package dlocker

type Options struct {
	Data        string
	AutoRenewal bool
	Reentrant   bool
	Fencing     bool
}

type Option func(opts *Options)

// 设置数据
func WithData(data string) Option {
	return func(opts *Options) {
		opts.Data = data
	}
}

// 自动续期
func WithAutoRenewal() Option {
	return func(opts *Options) {
		opts.AutoRenewal = true
	}
}

// WithFencing 启用 fencing token，默认关闭
func WithFencing() Option {
	return func(opts *Options) {
		opts.Fencing = true
	}
}

// 是否可重入，默认true(保持历史功能一致)
func WithReentrant(enable bool) Option {
	return func(opts *Options) {
		opts.Reentrant = enable
	}
}
