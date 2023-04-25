package tpl

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/zhiyunliu/glue/xdb"
)

type DBParam map[string]interface{}

func (p DBParam) Get(name string, ph Placeholder) (phName string, argVal interface{}) {
	val := p.GetVal(name)
	argName, phName := ph.Get(name)
	if tmpv, ok := val.(sql.NamedArg); ok {
		argVal = tmpv
		return
	}
	argVal = ph.BuildArgVal(argName, val)
	return
}

func (p DBParam) GetVal(name string) (val interface{}) {
	val, ok := p[name]
	if !ok {
		return nil
	}
	switch t := val.(type) {
	case sql.NamedArg:
		val = t
		return
	case *sql.NamedArg:
		val = *t
		return
	case driver.Valuer:
		val, _ = t.Value()
	case time.Time:
		val = t.Format(xdb.DateFormat)
	case *time.Time:
		if t != nil {
			val = t.Format(xdb.DateFormat)
		} else {
			val = nil
		}
	case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		val = strings.Trim(strings.Replace(fmt.Sprint(t), " ", ",", -1), "[]")
	case []string:
		val = "'" + strings.Join(t, "','") + "'"
	case driver.Value:
		val = t
	}
	return
}

func TransArgs(args []sql.NamedArg) []interface{} {
	result := make([]interface{}, len(args))
	for i := range args {
		result[i] = args[i]
	}
	return result
}
