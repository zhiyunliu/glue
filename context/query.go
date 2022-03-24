package context

import (
	"github.com/zhiyunliu/golibs/xtypes"
)

type Query interface {
	Get(name string) string
	SMap() xtypes.SMap
	Scan(obj interface{}) error
	String() string
}
