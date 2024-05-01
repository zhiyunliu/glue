package engine

type RouterOption interface {
	Apply(*RouterOptions)
}

type RouterOptions struct {
	Methods        []string
	ExcludeLogReq  bool
	ExcludeLogResp bool
}

// func (opts *RouterOptions) Merge(nopts *RouterOptions) *RouterOptions {
// 	ropts := &RouterOptions{}
// 	if nopts == nil {

// 	}

// 	return ropts
// }

type NormalRouterOption struct {
	callback func(*RouterOptions)
}

func (o *NormalRouterOption) Apply(opts *RouterOptions) {
	o.callback(opts)
}

func WithMethod(method ...string) RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.Methods = method
		},
	}
}

func WithExcludeLogReq() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.ExcludeLogReq = true
		},
	}
}
func WithExcludeLogResp() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.ExcludeLogResp = true
		},
	}
}
