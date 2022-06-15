package sqlserver

import (
	"database/sql"
	"fmt"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
)

//MssqlContext  模板
type MssqlContext struct {
	name    string
	prefix  string
	symbols tpl.Symbols
}

func New(name, prefix string) tpl.SQLTemplate {
	return &MssqlContext{
		name:    name,
		prefix:  prefix,
		symbols: newMssqlSymbols(),
	}
}

func (ctx *MssqlContext) Name() string {
	return ctx.name
}

//GetSQLContext 获取查询串
func (ctx *MssqlContext) GetSQLContext(template string, input map[string]interface{}) (query string, args []interface{}) {
	return tpl.AnalyzeTPLFromCache(ctx, template, input)
}

func (ctx *MssqlContext) Placeholder() tpl.Placeholder {
	index := 0
	f := func() string {
		index++
		return fmt.Sprint(ctx.prefix, index)
	}
	return f
}

func (ctx *MssqlContext) AnalyzeTPL(template string, input map[string]interface{}) (string, []string, []interface{}) {
	return tpl.DefaultAnalyze(ctx.symbols, template, input, ctx.Placeholder())
}

func getPlaceHolder(value interface{}, placeholder tpl.Placeholder) string {
	if arg, ok := value.(sql.NamedArg); ok {
		return "@" + arg.Name
	}
	return placeholder()
}

func newMssqlSymbols() tpl.Symbols {

	symbols := make(tpl.Symbols)
	symbols["@"] = func(input map[string]interface{}, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[fullKey]; ok {
			return ph
		}
		value := input[propName]
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
		} else {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, "")
		}
		item.NameCache[fullKey] = getPlaceHolder(value, item.Placeholder)
		return item.NameCache[fullKey]
	}

	symbols["$"] = func(input map[string]interface{}, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		value := input[propName]
		if !tpl.IsNil(value) {
			return fmt.Sprintf("%v", value)
		}
		return ""
	}

	symbols["&"] = func(input map[string]interface{}, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[fullKey]; ok {
			return ph
		}
		value := input[propName]
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf("and %s=%s", fullKey, item.Placeholder())
		}
		return ""
	}
	symbols["|"] = func(input map[string]interface{}, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[fullKey]; ok {
			return ph
		}
		value := input[propName]
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf("or %s=%s", fullKey, item.Placeholder())
		}
		return ""
	}
	return symbols
}
