package context

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
