package engine

type RouterOption func(opts *RouterOptions)

type RouterOptions struct {
	Methods           []string
	IgnoreLogRequest  bool
	IgnoreLogResponse bool
}

func WithMethod(method ...string) RouterOption {
	return func(opts *RouterOptions) {
		opts.Methods = method
	}
}

func WithIgnoreLogRequest() RouterOption {
	return func(opts *RouterOptions) {
		opts.IgnoreLogRequest = true
	}
}

func WithIgnoreLogResponse() RouterOption {
	return func(opts *RouterOptions) {
		opts.IgnoreLogResponse = true
	}
}
