package context

const XRequestId = "x-request-id"

type Request interface {
	GetMethod() string
	GetClientIP() string
	//Header() Header
	GetHeader(key string) string
	SetHeader(key, val string)
	Path() Path
	Query() Query
	Body() Body
}
