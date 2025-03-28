package context

type Request interface {
	GetMethod() string
	GetClientIP() string
	GetRemoteAddr() string
	RequestID() string
	ContentType() string
	Header() Header
	GetHeader(key string) string
	SetHeader(key, val string)
	Path() Path
	Query() Query
	Body() Body
	GetContentLength() int64
	GetImpl() interface{}
}
