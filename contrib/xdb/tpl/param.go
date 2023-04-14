package tpl

import (
	"database/sql/driver"
	"fmt"
	"strings"
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
	case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		return strings.Trim(strings.Replace(fmt.Sprint(t), " ", ",", -1), "[]")
	case []string:
		return "'" + strings.Join(t, "','") + "'"
	case driver.Value:
		val = t
		return
	}
	return
}
