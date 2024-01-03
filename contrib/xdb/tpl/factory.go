package tpl

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/xdb"
)

var (
	tpls map[string]SQLTemplate
)

const (
	TotalPattern = `[@]\{\w*[\.]?\w+\}|[&]\{\w*[\.]?\w+\}|[&]\{\w*[\.]?\w+\slike\s?\}|[&]\{\w*[\.]?\w+\s>\s?\}|[&]\{\w*[\.]?\w+\s>=\s?\}|[&]\{\w*[\.]?\w+\s<\s?\}|[&]\{\w*[\.]?\w+\s<=\s?\}|[\|]\{\w*[\.]?\w+\}`
	// ParamPattern   = `[@]\{\w*[\.]?\w+\}`
	// AndPattern     = `[&]\{\w*[\.]?\w+\}`
	// OrPattern      = `[\|]\{\w*[\.]?\w+\}`
	ReplacePattern = `\$\{\w*[\.]?\w+\}`

	SymbolAt  = "@"
	SymbolAnd = "&"
	SymbolOr  = "|"
)

// 符号回调函数
type SymbolCallback func(DBParam, string, *ReplaceItem) (string, xdb.MissError)
type SymbolMap interface {
	GetPattern() string
	//Store(name string, callback SymbolCallback)
	Register(Symbol) error
	Operator(Operator) error
	LoadOperator(oper string) (OperatorCallback, bool)
	LoadOrStore(name string, callback SymbolCallback) (loaded bool)
	Delete(name string)
	Load(name string) (SymbolCallback, bool)
	Clone() SymbolMap
}

type OperatorCallback func(string, string, string) string

type OperatorMap interface {
	LoadOrStore(name string, callback OperatorCallback) (loaded bool)
	Load(name string) (OperatorCallback, bool)
	Clone() OperatorMap
}

type Operator interface {
	Name() string
	Callback(string, string, string) string
}

type Symbol interface {
	Name() string
	GetPattern() string
	Callback(DBParam, string, *ReplaceItem) (string, xdb.MissError)
}

type Placeholder interface {
	Get(propName string) (argName string, phName string)
	NamedArg(name string) string
	BuildArgVal(argName string, val interface{}) interface{}
	Clone() Placeholder
}

// Template 模板上下文
type SQLTemplate interface {
	Name() string
	Placeholder() Placeholder
	GetSQLContext(tpl string, input map[string]interface{}) (query string, args []any, err error)
	AnalyzeTPL(tpl string, input map[string]interface{}, ph Placeholder) (sql string, item *ReplaceItem, err error)
	RegisterSymbol(symbol Symbol) error
	RegisterOperator(Operator) error
}

func init() {
	tpls = make(map[string]SQLTemplate)
}
func Register(tpl SQLTemplate) {
	if _, ok := tpls[tpl.Name()]; ok {
		log.Warnf("%s 注册被覆盖", tpl.Name())
	}
	tpls[tpl.Name()] = tpl
}

// GetDBTemplate 获取数据库上下文操作
func GetDBTemplate(name string) (SQLTemplate, error) {
	if v, ok := tpls[strings.ToLower(name)]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("不支持的数据库类型:%s", name)
}

// RegisterSymbol 给数据库注册语法解析
func RegisterSymbol(dbProto string, symbol Symbol) error {
	tmpl, err := GetDBTemplate(dbProto)
	if err != nil {
		return err
	}
	return tmpl.RegisterSymbol(symbol)
}

// RegisterSymbol 给数据库注册语法解析
func RegisterOperator(dbProto string, oper Operator) error {
	tmpl, err := GetDBTemplate(dbProto)
	if err != nil {
		return err
	}
	return tmpl.RegisterOperator(oper)
}
