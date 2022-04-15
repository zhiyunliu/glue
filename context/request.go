package context

const XRequestId = "x-request-id"
const XRemoteHeader = "x-remote-addr"

type Request interface {
	GetMethod() string
	GetClientIP() string
	Header() map[string]string
	GetHeader(key string) string
	SetHeader(key, val string)
	Path() Path
	Query() Query
	Body() Body
}
