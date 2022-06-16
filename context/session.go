package context

import (
	"context"

	"github.com/zhiyunliu/golibs/session"
)

type sidkey struct{}

func GetSid(ctx context.Context) string {
	sid, ok := ctx.Value(&sidkey{}).(string)
	if ok {
		sid = session.Create()
		WithSid(ctx, sid)
	}
	return sid

}

func WithSid(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, &sidkey{}, sid)
}
