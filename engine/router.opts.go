package engine

import (
	"sync/atomic"

	"github.com/zhiyunliu/glue/constants"
)

type RouterOption interface {
	Apply(*RouterOptions)
}

type AtomicValue[T any] struct {
	value atomic.Value
}

func (v *AtomicValue[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return v.value.CompareAndSwap(old, new)
}
func (v *AtomicValue[T]) Load() (val T) {
	return v.value.Load().(T)
}
func (v *AtomicValue[T]) Store(val T) {
	v.value.Store(val)
}
func (v *AtomicValue[T]) Swap(new T) (old T) {
	return v.value.Swap(new).((T))
}

func NewAtomicValue[T any](val T) *AtomicValue[T] {
	opts := &AtomicValue[T]{
		value: atomic.Value{},
	}
	opts.Store(val)
	return opts
}

type RouterOptions struct {
	Methods []string `json:"-"`
	// 排除请求日志
	ExcludeLogReq bool `json:"exclude_log_req"`
	// 排除响应日志
	ExcludeLogResp bool `json:"exclude_log_resp"`
	// 排除请求头
	ExcludeHeader bool `json:"exclude_header"`
	// 强制打印请求日志
	MandatoryLogReq bool `json:"mandatory_log_req"`
	// 强制打印响应日志
	MandatoryLogResp bool `json:"mandatory_log_resp"`
	// 打印请求源
	WithSource *bool `json:"with_source"`
	// 打印请求头
	WithHeaders []constants.HeaderGetter `json:"with_headers"`
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
