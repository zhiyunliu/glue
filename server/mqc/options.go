package mqc

import "net/http"

// Option 参数设置类型
type Option func(*options)

type options struct {
	addr, certFile, keyFile string
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
