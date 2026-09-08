package circuitbreaker

type Options struct {
	ConfigData []byte
}

type Option func(opts *Options)

// WithConfigData with the config data of circuit breaker.
func WithConfigData(config []byte) Option {
	return func(c *Options) {
		c.ConfigData = config
	}
}
