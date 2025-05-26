package constants

const (
	HeaderRequestId    = "X-Request-Id"
	HeaderRemoteHeader = "X-Remote-Addr"
	HeaderSourceIp     = "X-Src-Ip"
	HeaderSourceName   = "X-Src-Name"
)

const (
	HeaderXForwardedFor StrHeaderGetter = "X-Forwarded-For"
	HeaderAuthorization StrHeaderGetter = "Authorization"
	HeaderReferer       StrHeaderGetter = "Referer"
	HeaderAuthUserId    StrHeaderGetter = AUTH_USER_ID
)

var (
	DefaultHeaders = []HeaderGetter{HeaderXForwardedFor, HeaderReferer, HeaderAuthUserId}
)

type Header interface {
	Get(key string) string
}

type HeaderGetter interface {
	Key() string
	Get(headers Header) string
}

type StrHeaderGetter string

func (s StrHeaderGetter) Key() string {
	return string(s)
}

func (s StrHeaderGetter) Get(headers Header) string {
	if headers == nil {
		return ""
	}
	return headers.Get(string(s))
}
