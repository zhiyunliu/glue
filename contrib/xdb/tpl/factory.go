package tpl

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/log"
)

var (
	tpls map[string]SQLTemplate
)

const (
	TotalPattern = `[@]\{\w*[\.]?\w+\}|[&]\{\w*[\.]?\w+\}|[\|]\{\w*[\.]?\w+\}`
	// ParamPattern   = `[@]\{\w*[\.]?\w+\}`
	// AndPattern     = `[&]\{\w*[\.]?\w+\}`
	// OrPattern      = `[\|]\{\w*[\.]?\w+\}`
	ReplacePattern = `\$\{\w*[\.]?\w+\}`
)

type SymbolCallback func(DBParam, string, *ReplaceItem) string
type Symbols map[string]SymbolCallback
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
	GetSQLContext(tpl string, input map[string]interface{}) (query string, args []any)
	AnalyzeTPL(tpl string, input map[string]interface{}, ph Placeholder) (sql string, item *ReplaceItem)
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
