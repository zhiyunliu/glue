package xdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

type DbParamConverter interface {
	ToDbParam() map[string]any
}

type Placeholder interface {
	Get(propName string) (argName string, phName string)
	NamedArg(name string) string
	BuildArgVal(argName string, val interface{}) interface{}
	//Clone() Placeholder
}

// CheckIsNil 检查是否为空
var CheckIsNil func(input any) bool = DefaultIsNil

type DBParam map[string]any

// func (p DBParam) Get(name string, ph Placeholder) (phName string, argVal interface{}, err MissError) {
// 	val, err := p.GetVal(name)
// 	if err != nil {
// 		return
// 	}
// 	if tmpv, ok := val.(sql.NamedArg); ok {
// 		argVal = tmpv
// 		phName = ph.NamedArg(tmpv.Name)
// 		return
// 	}
// 	argName, phName := ph.Get(name)
// 	argVal = ph.BuildArgVal(argName, val)
// 	return
// }

func (p DBParam) GetVal(name string) (val interface{}, err MissError) {
	val, ok := p[name]
	if !ok {
		err = NewMissParamError(name, nil)
		return
	}
	switch t := val.(type) {
	case sql.NamedArg:
		val = t
		return
	case *sql.NamedArg:
		if t != nil {
			val = *t
		} else {
			val = nil
		}
		return
	case driver.Valuer:
		v, verr := t.Value()
		if verr != nil {
			val = nil
			// 处理 Value 方法返回的错误
			err = NewMissParamError(name, fmt.Errorf("[%s]failed to get value from driver.Valuer: %w", name, verr))
			return
		}
		val = v
		return
	case time.Time:
		val = t.Format(DateFormat)
		return
	case *time.Time:
		if t != nil {
			val = t.Format(DateFormat)
		} else {
			val = nil
		}
		return
	case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		val = t
		return
	case []string:
		val = t
		return
	case driver.Value:
		val = t
		return
	default:
		val = t
	}
	return

}

func TransArgs(args []sql.NamedArg) []any {
	result := make([]interface{}, len(args))
	for i := range args {
		result[i] = args[i]
	}
	return result
}

func DefaultIsNil(input interface{}) bool {
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
