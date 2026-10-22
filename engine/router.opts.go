package engine

import (
	"encoding/json"
	"sync/atomic"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/golibs/xreflect"
)

type RouterOption interface {
	Apply(*RouterOptions)
}

type AtomicValue[T any] struct {
	value    atomic.Value
	original T
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
		value:    atomic.Value{},
		original: xreflect.DeepCopy(val),
	}
	opts.Store(val)
	return opts
}

type HeaderGetters []constants.HeaderGetter

func (h *HeaderGetters) UnmarshalJSON(data []byte) (err error) {
	list := []string{}
	err = json.Unmarshal(data, &list)
	if err != nil {
		return err
	}

	*h = make(HeaderGetters, len(list))
	for i, key := range list {
		(*h)[i] = constants.StrHeaderGetter(key)
	}
	return nil
}

type RouterOptions struct {
	Methods []string `json:"-"`

	//重置
	Reset *bool `json:"reset"`

	SkipLog *bool `json:"skip_log"`
	// 排除请求日志
	ExcludeLogReq *bool `json:"exclude_log_req"`
	// 排除响应日志
	ExcludeLogResp *bool `json:"exclude_log_resp"`
	// 排除请求头
	ExcludeHeader *bool `json:"exclude_header"`
	// 强制打印请求日志
	MandatoryLogReq *bool `json:"mandatory_log_req"`
	// 强制打印响应日志
	MandatoryLogResp *bool `json:"mandatory_log_resp"`
	// 打印请求源
	WithSource *bool `json:"with_source"`
	// 打印请求头
	WithHeaders HeaderGetters `json:"with_headers"`
}

func (o RouterOptions) GetSkipLog() bool {
	return o.SkipLog != nil && *o.SkipLog
}

func (o RouterOptions) GetExcludeLogReq() bool {
	return o.ExcludeLogReq != nil && *o.ExcludeLogReq
}

func (o RouterOptions) GetExcludeLogResp() bool {
	return o.ExcludeLogResp != nil && *o.ExcludeLogResp
}

func (o RouterOptions) GetExcludeHeader() bool {
	return o.ExcludeHeader != nil && *o.ExcludeHeader
}

func (o RouterOptions) GetMandatoryLogReq() bool {
	return o.MandatoryLogReq != nil && *o.MandatoryLogReq
}
func (o RouterOptions) GetMandatoryLogResp() bool {
	return o.MandatoryLogResp != nil && *o.MandatoryLogResp
}

type NormalRouterOption struct {
	callback func(*RouterOptions)
}

func (o *NormalRouterOption) Apply(opts *RouterOptions) {
	o.callback(opts)
}

var _boolTrue = true

// WithMethod 设置方法
func WithMethod(method ...string) RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.Methods = method
		},
	}
}

// SkipLog 跳过日志
func SkipLog(skip bool) RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.SkipLog = &skip
		},
	}
}

// WithExcludeLogReq 排除打印请求日志
func WithExcludeLogReq() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.ExcludeLogReq = &_boolTrue
		},
	}
}

// WithExcludeLogResp 排除打印响应日志
func WithExcludeLogResp() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.ExcludeLogResp = &_boolTrue
		},
	}
}

// WithMandatoryLogReq 强制打印请求日志
func WithMandatoryLogReq() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.MandatoryLogReq = &_boolTrue
		},
	}
}

// WithMandatoryLogResp 强制打印响应日志
func WithMandatoryLogResp() RouterOption {
	return &NormalRouterOption{
		callback: func(opts *RouterOptions) {
			opts.MandatoryLogResp = &_boolTrue
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
