package context

import "github.com/zhiyunliu/velocity/libs/types"

type Header interface {
	Get(name string) string
	XMap() types.SMap
	Scan(obj interface{}) error
}
