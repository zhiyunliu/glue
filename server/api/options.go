package api

import (
	"net/http"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	addr, certFile, keyFile string
	handler                 http.Handler
	startedHook             func()
	endHook                 func()
}

func setDefaultOption() options {
	return options{
		addr: ":8080",
		handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
}

func WithEndHook(f func()) Option {
	return func(o *options) {
		o.endHook = f
	}
}

// WithStartedHook 设置启动回调函数
func WithStartedHook(f func()) Option {
	return func(o *options) {
		o.startedHook = f
	}
}

// WithAddr 设置addr
func WithAddr(s string) Option {
	return func(o *options) {
		o.addr = s
	}
}

// WithHandler 设置handler
func WithHandler(handler http.Handler) Option {
	return func(o *options) {
		o.handler = handler
	}
}
