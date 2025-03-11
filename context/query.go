package context

import (
	"net/url"

	"github.com/zhiyunliu/golibs/xtypes"
)

type Query interface {
	Get(name string) string
	Values() xtypes.SMap
	// Deprecated: use ScanTo instead
	Scan(obj interface{}) error
	ScanTo(obj interface{}) error
	String() string
	GetValues() url.Values
}
