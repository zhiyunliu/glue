package context

import (
	"context"

	"github.com/zhiyunliu/gel/log"
)

type Context interface {
	GetImpl() interface{}
	ServerType() string
	ResetContext(ctx context.Context)
	Context() context.Context
	Header(key string) string
	Request() Request
	Response() Response
	Log() log.Logger
	Close()
}
