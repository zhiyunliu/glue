package middleware

type MiddlewareBuilder interface {
	Build(data RawMessage) Middleware
	Name() string
}

// Middleware is HTTP/gRPC transport middleware.
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

func Resolve(m *Config) Middleware {
	xm, ok := _middlewareMap[m.Name]
	if !ok {
		return nil
	}
	return xm.Build(m.Data)
}
