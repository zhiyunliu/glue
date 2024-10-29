package tpl

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/xdb"
)

var (
	tpls               map[string]SQLTemplate
	DefaultPatternList []string = []string{
		`(@\{\w+(\.\w+)?\})`,
		`([&|\|](({(like\s+%?\w+(\.\w+)*%?}))|({\w+(\.\w+)?\s+like\s+%?\w+%?})))`,
		`([&|\|](({in\s+\w+(\.\w+)*(=\w+)?\})|({\w+(\.\w+)?\s+in\s+\w+})))`,
		`([&|\|](({(>|>=|<|<=)\s+\w+(\.\w+)?})|({(\w+\.)?\w+(>|>=|<|<=)\w+})))`,
	}
)

const (
	//TotalPattern = `(@\{\w+(\.\w+)?\})|([&|\|]\{like\s+%?\w+(\.\w+)*%?})|([&|\|]\{in\s+\w+(\.\w+)*(=\w+)?\})|([&|\|]\{((>|>=|<|<=)\s+)?\w+(\.\w+)*(=\w+)?\})`
	// ParamPattern   = `[@]\{\w*[\.]?\w+\}`
	// AndPattern     = `[&]\{\w*[\.]?\w+\}`
	// OrPattern      = `[\|]\{\w*[\.]?\w+\}`

	//替换
	ReplacePattern = `\$\{\w*[\.]?\w+\}`

	SymbolAt      = "@"
	SymbolAnd     = "&"
	SymbolOr      = "|"
	SymbolReplace = "$"
)

// 符号回调函数
type SymbolCallback func(SymbolMap, xdb.DBParam, string, *ReplaceItem) (string, xdb.MissError)
type SymbolMap interface {
	GetPattern() string
	RegisterSymbol(Symbol) error
	RegisterOperator(Operator) error
	StoreSymbol(name string, callback SymbolCallback)
	LoadSymbol(name string) (SymbolCallback, bool)
	Delete(name string)
	Clone() SymbolMap
}

type OperatorCallback func(string, string, string) string

type OperatorMap interface {
	Store(name string, callback OperatorCallback)
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
	Callback(SymbolMap, xdb.DBParam, string, *ReplaceItem) (string, xdb.MissError)
}

// Template 模板上下文
type SQLTemplate interface {
	Name() string
	Placeholder() xdb.Placeholder
	GetSQLContext(tpl string, input map[string]interface{}) (query string, args []any, err error)
	AnalyzeTPL(tpl string, input map[string]interface{}, ph xdb.Placeholder) (sql string, item *ReplaceItem, err error)
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
