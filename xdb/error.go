package xdb

import (
	"fmt"
	"reflect"
	"strings"
)

type DbError interface {
	Error() string
	Inner() error
	SQL() string
	Args() []interface{}
}

type MissParamError interface {
	Error() string
	ParamName() string
}

type MissParamsError interface {
	Error() string
	ParamList() []string
}

type PanicError interface {
	Error() string
	StackTrace() string
}

type InvalidArgTypeError struct {
	Type reflect.Type
}

func (e InvalidArgTypeError) Error() string {
	if e.Type == nil {
		return "result(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "non-pointer " + e.Type.String()
	}
	return "nil " + e.Type.String()
}

type xDBError struct {
	innerError error
	sql        string
	stackTrace string
	args       []interface{}
}

type xMissParamError struct {
	paramName string
}

func (e xMissParamError) Error() string {
	return fmt.Sprintf("SQL缺少参数:[%s]", e.paramName)
}

func (e xMissParamError) ParamName() string {
	return e.paramName
}

func NewMissParamError(name string) MissParamError {
	return &xMissParamError{
		paramName: name,
	}
}

type xMissParamsError struct {
	paramList []string
}

func (e xMissParamsError) Error() string {
	return fmt.Sprintf("SQL缺少参数:[%s]", strings.Join(e.paramList, ","))
}

func (e xMissParamsError) ParamList() []string {
	return e.paramList
}

func NewMissParamsError(errList ...MissParamError) MissParamsError {
	paramList := make([]string, len(errList))
	for i := range errList {
		paramList[i] = errList[i].ParamName()
	}
	return &xMissParamsError{
		paramList: paramList,
	}
}

func NewError(err error, execSql string, args []interface{}) error {
	return &xDBError{
		innerError: err,
		sql:        execSql,
		args:       args,
	}
}

func (e xDBError) Error() string {
	return e.innerError.Error()
}

func (e xDBError) Inner() error {
	return e.innerError
}

func (e xDBError) SQL() string {
	return e.sql
}

func (e xDBError) Args() []interface{} {
	return e.args
}

func (e xDBError) StackTrace() string {
	return e.stackTrace
}

func NewPanicError(err error, strace string) error {
	return &xDBError{
		innerError: err,
		stackTrace: strace,
	}
}
