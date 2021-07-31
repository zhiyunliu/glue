package logger

import "context"

type loggerKey struct{}

func FromContext(ctx context.Context) (*Wrapper, bool) {
	l, ok := ctx.Value(&loggerKey{}).(*Wrapper)
	return l, ok
}

func NewContext(ctx context.Context, l *Wrapper) context.Context {
	return context.WithValue(ctx, &loggerKey{}, l)
}
