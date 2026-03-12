package context

import (
	"net/url"

	"github.com/zhiyunliu/golibs/xtypes"
)

type Query interface {
	Get(name string) string
	Values() xtypes.SMap
	ScanTo(obj interface{}) error
	String() string
	GetValues() url.Values
}
