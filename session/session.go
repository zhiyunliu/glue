package session

import (
	"context"
)

type sessionKey struct{}

func FromContext(ctx context.Context) (string, bool) {
	l, ok := ctx.Value(&sessionKey{}).(string)
	return l, ok
}

func WithContext(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, &sessionKey{}, sid)
}
