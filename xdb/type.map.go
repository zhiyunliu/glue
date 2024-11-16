package xdb

import (
	"fmt"
	"reflect"

	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/xtypes"
)

var typeMap map[string]map[string]reflect.Type

func init() {
	typeMap = make(map[string]map[string]reflect.Type)
	RegisterDbType("DECIMAL", reflect.TypeOf((*xtypes.Decimal)(nil)).Elem())
	RegisterDbType("MONEY", reflect.TypeOf((*xtypes.Decimal)(nil)).Elem())
}

func GetDbType(proto, dbtype string) (reflect.Type, bool) {
	if tmp, ok := typeMap[dbtype]; ok {
		t, ok := tmp[proto]
		if ok {
			return t, ok
		}
		t, ok = tmp["*"]
		return t, ok
	}

	return nil, false
}

func RegisterDbType(dbType string, reflectType reflect.Type) (err error) {
	return RegisterProtoDbType("*", dbType, reflectType)
}

func RegisterProtoDbType(proto, dbType string, reflectType reflect.Type) (err error) {
	if global.IsRunning() {
		err = fmt.Errorf("MicroApp已启动,不能再注册类型.请在MicroApp.Start之前注册")
		return
	}

	defer func() {
		if obj := recover(); obj != nil {
			err = obj.(error)
		}
	}()
	if tmp, ok := typeMap[dbType]; ok {
		tmp[proto] = reflectType
	} else {
		tmp = make(map[string]reflect.Type)
		tmp[proto] = reflectType
		typeMap[dbType] = tmp
	}
	return
}
