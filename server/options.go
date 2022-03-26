package server

type Option func(*options)

type options struct {
	SrvType         string
	RequestDecoder  DecodeRequestFunc  //:          server.DefaultRequestDecoder,
	ResponseEncoder EncodeResponseFunc //:          server.DefaultResponseEncoder,
	ErrorEncoder    EncodeErrorFunc    //:          server.DefaultErrorEncoder,
}

func setDefaultOptions() *options {
	return &options{
		SrvType:         "api",
		RequestDecoder:  DefaultRequestDecoder,
		ResponseEncoder: DefaultResponseEncoder,
		ErrorEncoder:    DefaultErrorEncoder,
	}
}

func WithSrvType(srvType string) Option {
	return func(o *options) {
		o.SrvType = srvType
	}
}

func WithRequestDecoder(requestDecoder DecodeRequestFunc) Option {
	return func(o *options) {
		o.RequestDecoder = requestDecoder
	}
}

func WithResponseEncoder(responseEncoder EncodeResponseFunc) Option {
	return func(o *options) {
		o.ResponseEncoder = responseEncoder
	}
}

func WithErrorEncoder(errorEncoder EncodeErrorFunc) Option {
	return func(o *options) {
		o.ErrorEncoder = errorEncoder
	}
}
