package sqlserver

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
)

//MssqlContext  模板
type MssqlContext struct {
	name    string
	prefix  string
	symbols tpl.Symbols
}

type mssqlPlaceHolder struct {
	ctx *MssqlContext
	idx int
}

func (ph *mssqlPlaceHolder) Get() string {
	ph.idx++
	return fmt.Sprint(ph.ctx.prefix, ph.idx)
}

func (ph *mssqlPlaceHolder) Clone() tpl.Placeholder {
	return &mssqlPlaceHolder{
		idx: ph.idx,
		ctx: ph.ctx,
	}
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
	return tpl.AnalyzeTPLFromCache(ctx, template, input, ctx.Placeholder())
}

func (ctx *MssqlContext) Placeholder() tpl.Placeholder {
	return &mssqlPlaceHolder{ctx: ctx, idx: 0}
}

func (ctx *MssqlContext) AnalyzeTPL(template string, input map[string]interface{}, ph tpl.Placeholder) (string, []string, []interface{}) {
	return tpl.DefaultAnalyze(ctx.symbols, template, input, ph)
}

func getPlaceHolder(value interface{}, placeholder tpl.Placeholder) string {
	if arg, ok := value.(sql.NamedArg); ok {
		return "@" + arg.Name
	}
	if arg, ok := value.(*sql.NamedArg); ok {
		return "@" + arg.Name
	}
	return placeholder.Get()
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

	symbols["&"] = func(input map[string]interface{}, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		value := input[propName]
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf("and %s=%s", fullKey, item.Placeholder.Get())
		}
		return ""
	}
	symbols["|"] = func(input map[string]interface{}, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		value := input[propName]
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf("or %s=%s", fullKey, item.Placeholder.Get())
		}
		return ""
	}
	return symbols
}

func (ctx *MssqlContext) HandleAndSymbols(template string, input map[string]interface{}, ph tpl.Placeholder) (sql string, values []interface{}, exists bool) {
	word := regexp.MustCompile(tpl.AndPattern)
	item := &tpl.ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ph,
	}
	symbols := ctx.symbols
	exists = false
	//变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(template, func(s string) string {
		exists = true
		symbol := s[:1]
		fullKey := s[2 : len(s)-1]
		callback, ok := symbols[symbol]
		if !ok {
			return s
		}
		return callback(input, fullKey, item)
	})

	return sql, item.Values, exists
}

func (ctx *MssqlContext) HandleOrSymbols(template string, input map[string]interface{}, ph tpl.Placeholder) (sql string, values []interface{}, exists bool) {
	word := regexp.MustCompile(tpl.OrPattern)
	item := &tpl.ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ph,
	}
	symbols := ctx.symbols
	exists = false
	//变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(template, func(s string) string {
		exists = true
		symbol := s[:1]
		fullKey := s[2 : len(s)-1]
		callback, ok := symbols[symbol]
		if !ok {
			return s
		}
		return callback(input, fullKey, item)
	})

	return sql, item.Values, exists
}
