package redis

type Options struct {
	Addrs        []string `json:"addrs,omitempty"  valid:"required" `
	Username     string   `json:"username,omitempty" `
	Password     string   `json:"password,omitempty" `
	DbIndex      uint     `json:"db,omitempty"`
	DialTimeout  uint     `json:"dial_timeout,omitempty"`
	ReadTimeout  uint     `json:"read_timeout,omitempty"`
	WriteTimeout uint     `json:"write_timeout,omitempty" `
	PoolSize     uint     `json:"pool_size,omitempty"`
}

type Option func(opts *Options)

func WithAddrs(addrs ...string) Option {
	return func(opts *Options) {
		opts.Addrs = addrs
	}
}
func WithUsername(username string) Option {
	return func(opts *Options) {
		opts.Username = username
	}
}
func WithPassword(password string) Option {
	return func(opts *Options) {
		opts.Password = password
	}
}

func WithDbIndex(index uint) Option {
	return func(opts *Options) {
		opts.DbIndex = index
	}
}
func WithPoolSize(poolSize uint) Option {
	return func(opts *Options) {
		opts.PoolSize = poolSize
	}
}
func WithReadTimeout(timeout uint) Option {
	return func(opts *Options) {
		opts.ReadTimeout = timeout
	}
}
func WithWriteTimeout(timeout uint) Option {
	return func(opts *Options) {
		opts.WriteTimeout = timeout
	}
}

func WithDialTimeout(timeout uint) Option {
	return func(opts *Options) {
		opts.DialTimeout = timeout
	}
}
