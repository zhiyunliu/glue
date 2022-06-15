package context

import (
	"context"

	"github.com/zhiyunliu/glue/log"
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
	Meta() map[string]interface{}
	Log() log.Logger
	Close()
}
