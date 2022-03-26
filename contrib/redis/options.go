package redis

type options struct {
	Addrs        []string `json:"addrs,omitempty"  valid:"required" `
	Password     string   `json:"password,omitempty" `
	DbIndex      uint     `json:"db,omitempty"`
	DialTimeout  uint     `json:"dial_timeout,omitempty"`
	ReadTimeout  uint     `json:"read_timeout,omitempty"`
	WriteTimeout uint     `json:"write_timeout,omitempty" `
	PoolSize     uint     `json:"pool_size,omitempty"`
}

type Option func(opts *options)

func WithAddrs(addrs ...string) Option {
	return func(opts *options) {
		opts.Addrs = addrs
	}
}

func WithPassword(password string) Option {
	return func(opts *options) {
		opts.Password = password
	}
}

func WithDbIndex(index uint) Option {
	return func(opts *options) {
		opts.DbIndex = index
	}
}
func WithPoolSize(poolSize uint) Option {
	return func(opts *options) {
		opts.PoolSize = poolSize
	}
}
func WithReadTimeout(timeout uint) Option {
	return func(opts *options) {
		opts.ReadTimeout = timeout
	}
}
func WithWriteTimeout(timeout uint) Option {
	return func(opts *options) {
		opts.WriteTimeout = timeout
	}
}

func WithDialTimeout(timeout uint) Option {
	return func(opts *options) {
		opts.DialTimeout = timeout
	}
}
