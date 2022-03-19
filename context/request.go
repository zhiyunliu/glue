package context

type Request interface {
	GetMethod() string
	GetClientIP() string
	Header(key string) string
	//UserInfo() xtypes.XMap
	Path() Path
	Query() Query
	Body() Body
}
