package engine

import (
	"context"
)

type Request interface {
	Context() context.Context
	WithContext(context.Context)
	GetName() string
	GetService() string
	GetMethod() string
	Params() map[string]string
	GetHeader() map[string]string
	Body() []byte
	GetRemoteAddr() string
}
