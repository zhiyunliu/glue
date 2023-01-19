package tpl

import (
	"database/sql/driver"
	"time"

	"github.com/zhiyunliu/glue/xdb"
)

type DBParam map[string]interface{}

func (p DBParam) Get(name string) (val interface{}) {
	val = p[name]
	switch t := val.(type) {
	case driver.Valuer:
		val, _ = t.Value()
		return
	case time.Time:
		val = t.Format(xdb.DateFormat)
		return
	case *time.Time:
		if t != nil {
			val = t.Format(xdb.DateFormat)
			return
		}
		val = ""
	case driver.Value:
		val = t
		return
	}
	return
}
