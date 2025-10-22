package xdb

import "reflect"

type StmtDbTypeHandler interface {
	Name() string
	//args:a=b => [a,b]
	//Handle(param any, args []string) any
	Handle(fieldName string, param any, fv reflect.Value, args []string) (any, error)
}

type StmtDbTypeProcessor interface {
	// RegistHandler 注册表达式匹配器
	RegistHandler(handler ...StmtDbTypeHandler)
	//Process(param any, tagOpts TagOptions) any
	Process(fieldName string, param any, fv reflect.Value, tagOpts TagOptions) (any, error)
}

var (
	// 新建一个请求参数处理器
	NewStmtDbTypeProcessor func(matchers ...StmtDbTypeHandler) StmtDbTypeProcessor
)
