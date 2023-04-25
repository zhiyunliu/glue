package internal

import (
	"reflect"

	"github.com/zhiyunliu/golibs/xtypes"
)

var typeMap map[string]reflect.Type

func init() {
	typeMap = make(map[string]reflect.Type)
	typeMap["DECIMAL"] = reflect.TypeOf((*xtypes.Decimal)(nil)).Elem()
	typeMap["MONEY"] = reflect.TypeOf((*xtypes.Decimal)(nil)).Elem()
}

func getDbType(dbtype string) (reflect.Type, bool) {
	t, ok := typeMap[dbtype]
	return t, ok
}

func RegisterDbType(dbType string, reflectType reflect.Type) (err error) {
	defer func() {
		if obj := recover(); obj != nil {
			err = obj.(error)
		}
	}()
	typeMap[dbType] = reflectType
	return
}
