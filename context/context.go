package context

type Context interface {
	Header(key string) string
	Request() Request
	Response() Response
	Log()
	Close()
	Trace(...interface{})
}
