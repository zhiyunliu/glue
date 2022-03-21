package context

import "github.com/zhiyunliu/velocity/extlib/xtypes"

type Query interface {
	Get(name string) string
	XMap() xtypes.SMap
	Scan(obj interface{}) error
	String() string
}
