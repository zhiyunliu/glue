package context

import "github.com/zhiyunliu/golibs/xtypes"

type Header interface {
	Get(name string) string
	Set(name, val string)
	Scan(obj interface{}) error
	Values() xtypes.SMap
}
