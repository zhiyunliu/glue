package constants

const (
	HeaderRequestId    = "X-Request-Id"
	HeaderRemoteHeader = "X-Remote-Addr"
	HeaderSourceIp     = "X-Src-Ip"
	HeaderSourceName   = "X-Src-Name"
)

const (
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderAuthorization = "Authorization"
	HeaderReferer       = "Referer"
)

var (
	DefaultHeaders = []string{HeaderXForwardedFor, HeaderReferer}
)
