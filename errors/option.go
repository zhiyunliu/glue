package errors

// WithSubCode 设置子错误码
func WithSubCode(subCode string) Option {
	return func(e *xError) {
		e.SubCode = subCode
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
