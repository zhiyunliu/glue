package dlocker

import "time"

type Options struct {
	Data                 string
	AutoRenewal          bool
	Reentrant            bool
	Fencing              bool
	Timeout              int
	TimeoutCheckInterval time.Duration
}

type Option func(opts *Options)

// WithTimeout 设置获取锁的最长等待时间，单位秒；0 表示只尝试一次。
// 负数或无法转换为时间间隔的值会在 Acquire 时返回参数错误。
func WithTimeout(timeout int) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

// WithTimeoutCheckInterval 设置超时检查间隔，默认 100ms
func WithTimeoutCheckInterval(timeoutCheckInterval time.Duration) Option {
	return func(opts *Options) {
		opts.TimeoutCheckInterval = timeoutCheckInterval
	}
}

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
