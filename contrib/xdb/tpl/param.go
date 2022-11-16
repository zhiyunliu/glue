package tpl

import (
	"time"

	"github.com/zhiyunliu/glue/xdb"
)

type DBParam map[string]interface{}

func (p DBParam) Get(name string) (val interface{}) {
	val = p[name]
	switch t := val.(type) {
	case time.Time:
		val = t.Format(xdb.DateFormat)
	case *time.Time:
		if t != nil {
			val = t.Format(xdb.DateFormat)
			return
		}
		val = ""
	}
	return
}
