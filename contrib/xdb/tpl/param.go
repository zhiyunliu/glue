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

func (p DBParam) Get(name string, ph Placeholder) (phName string, argVal interface{}, err xdb.MissParamError) {
	val, err := p.GetVal(name)
	if err != nil {
		return
	}
	if tmpv, ok := val.(sql.NamedArg); ok {
		argVal = tmpv
		phName = ph.NamedArg(tmpv.Name)
		return
	}
	argName, phName := ph.Get(name)
	argVal = ph.BuildArgVal(argName, val)
	return
}

func (p DBParam) GetVal(name string) (val interface{}, err xdb.MissParamError) {
	val, ok := p[name]
	if !ok {
		err = xdb.NewMissParamError(name)
		return
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
