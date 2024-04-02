package rabbit

import "github.com/zhiyunliu/glue/queue"

type options struct {
	queue.Options
	Addr        string `json:"addr,omitempty"  valid:"required" `
	VirtualHost string `json:"virtual_host,omitempty"`
	//BindKey      string `json:"bind_key,omitempty"`
	Exchange     string `json:"exchange,omitempty"`
	ExchangeType string `json:"exchange_type,omitempty"`
	ConnName     string `json:"conn_name,omitempty"`

	//Properties map[string]any `json:"properties"`
}

type Option func(opts *options)

func WithAddr(addr string) Option {
	return func(opts *options) {
		opts.Addr = addr
	}
}

func WithVirtualHost(virtualHost string) Option {
	return func(opts *options) {
		opts.VirtualHost = virtualHost
	}
}

func WithExchange(exchange string) Option {
	return func(opts *options) {
		opts.Exchange = exchange
	}
}

func WithExchangeType(exchangeType string) Option {
	return func(opts *options) {
		opts.ExchangeType = exchangeType
	}
}
