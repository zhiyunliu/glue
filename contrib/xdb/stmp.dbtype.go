package xdb

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/zhiyunliu/glue/xdb"
)

type StmtDbTypeOutputHandler struct {
}

var _ xdb.StmtDbTypeHandler = (*StmtDbTypeOutputHandler)(nil)

func (h *StmtDbTypeOutputHandler) Name() string {
	return "output"
}
func (h *StmtDbTypeOutputHandler) Handle(fieldName string, _ any, fv reflect.Value, _ []string) (any, error) {
	if !fv.CanSet() {
		return nil, fmt.Errorf("字段[%s]作为output参数不能被设置.请使用指针传递SQL参数", fieldName)
	}
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return sql.Named(fieldName, sql.Out{Dest: fv.Interface()}), nil
	}
	return sql.Named(fieldName, sql.Out{Dest: fv.Addr().Interface()}), nil
}
