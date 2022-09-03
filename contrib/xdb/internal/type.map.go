package internal

import (
	"github.com/zhiyunliu/golibs/xtypes"
	"reflect"
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
