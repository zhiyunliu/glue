package context

import "github.com/zhiyunliu/velocity/extlib/xtypes"

type Request interface {
	GetMethod() string
	GetClientIP() string
	Headers() Header
	Header(key string) string
	UserInfo() xtypes.XMap
	Path() Path
	Query() Query
	Body() Body
}
