package engine

import "github.com/zhiyunliu/glue/constants"

type RouterOption interface {
	Apply(*RouterOptions)
}

type RouterOptions struct {
	Methods []string
	// 排除请求日志
	ExcludeLogReq bool
	// 排除响应日志
	ExcludeLogResp bool
	// 强制打印请求日志
	MandatoryLogReq bool
	// 强制打印响应日志
	MandatoryLogResp bool
	// 打印请求头
	WithHeaders []constants.HeaderGetter //打印请求头
	// 打印请求源
	WithSource *bool //打印请求源
}

type NormalRouterOption struct {
	callback func(*RouterOptions)
}

func (o *NormalRouterOption) Apply(opts *RouterOptions) {
	o.callback(opts)
}

// WithMethod 设置方法
func WithMethod(method ...string) RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.Methods = method
		},
	}
}

// WithExcludeLogReq 排除打印请求日志
func WithExcludeLogReq() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.ExcludeLogReq = true
		},
	}
}

// WithExcludeLogResp 排除打印响应日志
func WithExcludeLogResp() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.ExcludeLogResp = true
		},
	}
}

// WithMandatoryLogReq 强制打印请求日志
func WithMandatoryLogReq() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.MandatoryLogReq = true
		},
	}
}

// WithMandatoryLogResp 强制打印响应日志
func WithMandatoryLogResp() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.MandatoryLogResp = true
		},
	}
}

// WithPrintHeaders 指定打印的请求头
func WithPrintHeaders(keys ...constants.HeaderGetter) RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.WithHeaders = keys
		},
	}
}

// WithPrintRequestBody 启用打印请求源
func WithPrintSource(include bool) RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.WithSource = &include
		},
	}
}
