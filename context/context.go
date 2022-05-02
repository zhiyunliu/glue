package context

import (
	"context"

	"github.com/zhiyunliu/gel/log"
)

type Context interface {
	GetImpl() interface{}
	ServerType() string
	ServerName() string
	ResetContext(ctx context.Context)
	Context() context.Context
	Header(key string) string
	Request() Request
	Bind(interface{}) error
	Response() Response
	Log() log.Logger
	Close()
}
