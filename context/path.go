package context

import (
	"net/url"

	"github.com/zhiyunliu/velocity/libs/types"
)

type Path interface {
	GetURL() *url.URL
	FullPath() string
	Params() types.SMap
}
