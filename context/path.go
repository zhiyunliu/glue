package context

import (
	"net/url"

	"github.com/zhiyunliu/golibs/xtypes"
)

type Path interface {
	GetURL() *url.URL
	FullPath() string
	Params() xtypes.SMap
}
