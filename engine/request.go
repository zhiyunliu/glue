package engine

import (
	"context"
	"net/url"
)

type Request interface {
	Context() context.Context
	WithContext(context.Context)
	GetName() string
	//GetService() string
	GetURL() *url.URL
	GetMethod() string
	Params() map[string]string
	GetHeader() map[string]string
	Body() []byte
	GetRemoteAddr() string
}
