package alloter

import "context"

type IRequest interface {
	Context() context.Context
	WithContext(context.Context) IRequest
	GetName() string
	GetService() string
	GetMethod() string
	Params() map[string]string
	GetHeader() map[string]string
	Body() []byte
	GetRemoteAddr() string
}
