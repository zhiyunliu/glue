package errors

type Option func(*xError)

// WithSubCode 设置子错误码
func WithSubCode(subCode string) Option {
	return func(e *xError) {
		e.SubCode = subCode
	}
}

// WithData 设置数据
func WithData(data any) Option {
	return func(e *xError) {
		e.Data = data
	}
}

// WithErrData 设置错误数据

func WithErrData(errData map[string]any) Option {
	return func(e *xError) {
		e.ErrData = errData
	}
}

// WithMetadata 添加元数据

func WithInnerErr(inner error) Option {
	return func(e *xError) {
		e.innerErr = inner
	}
}
func WithStatusCode(statusCode int) Option {
	return func(e *xError) {
		e.statusCode = statusCode
	}
}
