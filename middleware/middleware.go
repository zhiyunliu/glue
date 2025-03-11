package middleware

type MiddlewareBuilder interface {
	Build(data *Config) (Middleware, error)
	Name() string
}

// Middleware is HTTP/gRPC/Mqc/Cron transport middleware.
type Middleware func(Handler) Handler

// Chain returns a Middleware that specifies the chained handler for endpoint.
func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}

var _middlewareMap = map[string]MiddlewareBuilder{}

func Registry(x MiddlewareBuilder) {
	_middlewareMap[x.Name()] = x
}

func Resolve(m *Config) (Middleware, error) {
	xm, ok := _middlewareMap[m.Name]
	if !ok {
		return nil, nil
	}
	return xm.Build(m)
}

// 生成中间件列表
func BuildMiddlewareList(cfglist []Config) (midwares []Middleware, err error) {
	for _, m := range cfglist {
		midware, ierr := Resolve(&m)
		if ierr != nil {
			err = ierr
			return
		}
		if midware == nil {
			continue
		}
		midwares = append(midwares, midware)
	}
	return midwares, nil
}
