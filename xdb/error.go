package xdb

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	MissTypeParam  = "param"
	MissTypeOper   = "oper"
	MissTypeSymbol = "symbol"
)

type DbError interface {
	Error() string
	Inner() error
	SQL() string
	Args() []interface{}
}

type MissError interface {
	Error() string
	Type() string
	Name() string
}

type MissListError interface {
	Error() string
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
	innerErr  error
}

func (e xMissParamError) Error() string {
	if e.innerErr != nil {
		return e.Error()
	}
	return fmt.Sprintf("SQL缺少参数:[%s]", e.paramName)
}

func (e xMissParamError) Name() string {
	return e.paramName
}

func (e xMissParamError) Type() string {
	return MissTypeParam
}

func NewMissParamError(name string, innerErr error) MissError {
	return &xMissParamError{
		paramName: name,
		innerErr:  innerErr,
	}
}

type xMissOperError struct {
	operName string
}

func (e xMissOperError) Error() string {
	return fmt.Sprintf("缺少Operator定义:[%s]", e.operName)
}

func (e xMissOperError) Name() string {
	return e.operName
}

func (e xMissOperError) Type() string {
	return MissTypeOper
}

func NewMissOperError(name string) MissError {
	return &xMissOperError{
		operName: name,
	}
}

type xMissPropError struct {
	propName string
}

func (e xMissPropError) Error() string {
	return fmt.Sprintf("缺少Property定义:[%s]", e.propName)
}

func (e xMissPropError) Name() string {
	return e.propName
}

func (e xMissPropError) Type() string {
	return MissTypeOper
}

func NewMissPropError(name string) MissError {
	return &xMissPropError{
		propName: name,
	}
}

type xMissSymbolError struct {
	propName string
}

func (e xMissSymbolError) Error() string {
	return fmt.Sprintf("缺少Property定义:[%s]", e.propName)
}

func (e xMissSymbolError) Name() string {
	return e.propName
}

func (e xMissSymbolError) Type() string {
	return MissTypeSymbol
}

func NewMissSymbolError(name string) MissError {
	return &xMissSymbolError{
		propName: name,
	}
}

type xMissParamsError struct {
	paramList  []string
	operList   []string
	symbolList []string
	otherList  []string
}

func (e xMissParamsError) Error() string {
	msgList := []string{}
	if len(e.paramList) > 0 {
		msgList = append(msgList, fmt.Sprintf("SQL缺少参数:[%s]", strings.Join(e.paramList, ",")))
	}
	if len(e.operList) > 0 {
		msgList = append(msgList, fmt.Sprintf("缺少Operator定义:[%s]", strings.Join(e.operList, ",")))
	}
	if len(e.symbolList) > 0 {
		msgList = append(msgList, fmt.Sprintf("缺少Symbol定义:[%s]", strings.Join(e.symbolList, ",")))
	}
	if len(e.otherList) > 0 {
		msgList = append(msgList, fmt.Sprintf("缺少类型:[%s]", strings.Join(e.otherList, ",")))
	}
	return strings.Join(msgList, "\r\n")
}

func NewMissListError(errList ...MissError) MissListError {
	paramList := []string{}
	operList := []string{}
	symbolList := []string{}
	otherList := []string{}

	for i := range errList {
		if strings.EqualFold(errList[i].Type(), MissTypeParam) {
			paramList = append(paramList, errList[i].Name())
		} else if strings.EqualFold(errList[i].Type(), MissTypeOper) {
			operList = append(operList, errList[i].Name())
		} else if strings.EqualFold(errList[i].Type(), MissTypeSymbol) {
			symbolList = append(symbolList, errList[i].Name())
		} else {
			otherList = append(otherList, fmt.Sprintf("%s:%s", errList[i].Type(), errList[i].Name()))
		}
	}
	return &xMissParamsError{
		paramList:  paramList,
		operList:   operList,
		symbolList: symbolList,
		otherList:  otherList,
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
