package redis

type Options struct {
	Addrs        []string `json:"addrs,omitempty"  valid:"required" `
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

func WithPassword(password string) Option {
	return func(opts *Options) {
		opts.Password = password
	}
}
