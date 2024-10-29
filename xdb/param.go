package xdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type DbParamConverter interface {
	ToDbParam() map[string]any
}

type Placeholder interface {
	Get(propName string) (argName string, phName string)
	NamedArg(name string) string
	BuildArgVal(argName string, val interface{}) interface{}
	Clone() Placeholder
}

type DBParam map[string]any

func (p DBParam) Get(name string, ph Placeholder) (phName string, argVal interface{}, err MissError) {
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

func (p DBParam) GetVal(name string) (val interface{}, err MissError) {
	val, ok := p[name]
	if !ok {
		err = NewMissParamError(name)
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
		val = t.Format(DateFormat)
	case *time.Time:
		if t != nil {
			val = t.Format(DateFormat)
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

func IsNil(input interface{}) bool {
	if input == nil {
		return true
	}

	if arg, ok := input.(sql.NamedArg); ok {
		input = arg.Value
	}
	if arg, ok := input.(*sql.NamedArg); ok {
		input = arg.Value
	}
	if input == nil {
		return true
	}
	if fmt.Sprintf("%v", input) == "" {
		return true
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Ptr {
		return rv.IsNil()
	}
	return false
}
