package context

import "github.com/zhiyunliu/golibs/xtypes"

type Header interface {
	Get(name string) string
	Set(name, val string)
	Values() map[string]string
	Keys() []string
}

type MapHeader xtypes.SMap

func (m MapHeader) Values() map[string]string {
	return m
}
