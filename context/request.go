package context

type Request interface {
	GetMethod() string
	Headers() Header
	Header(key string) string
	UserInfo() UserInfo
	Path() Path
	Query() Query
	Body() Body
}
