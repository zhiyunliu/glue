package context

import (
	"context"

	"github.com/zhiyunliu/velocity/log"
)

type Context interface {
	ResetContext(ctx context.Context)
	Context() context.Context
	Header(key string) string
	Request() Request
	Response() Response
	Log() log.Logger
	Close()
}
