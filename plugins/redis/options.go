package redis

type Options struct {
	Addrs        []string `json:"addrs,omitempty"  valid:"required" `
	Password     string   `json:"password,omitempty" `
	DbIndex      int      `json:"db,omitempty"`
	DialTimeout  int      `json:"dial_timeout,omitempty"`
	ReadTimeout  int      `json:"read_timeout,omitempty"`
	WriteTimeout int      `json:"write_timeout,omitempty" `
	PoolSize     int      `json:"pool_size,omitempty"`
}

type Option func(opts *Options)

func WithAddrs(addrs ...string) Option {
	return func(opts *Options) {
		opts.Addrs = addrs
	}
}

func WithPassword(password string) Option {
	return func(opts *Options) {
		opts.Password = password
	}
}
