package sqlserver

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
)

// MssqlContext  模板
type MssqlContext struct {
	name    string
	prefix  string
	symbols tpl.Symbols
}

type mssqlPlaceHolder struct {
	ctx *MssqlContext
}

func (ph *mssqlPlaceHolder) Get(propName string) (argName, phName string) {
	argName = fmt.Sprint(ph.ctx.prefix, propName)
	phName = "@" + argName
	return
}

func (ph *mssqlPlaceHolder) Clone() tpl.Placeholder {
	return &mssqlPlaceHolder{
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

// GetSQLContext 获取查询串
func (ctx *MssqlContext) GetSQLContext(template string, input map[string]interface{}) (query string, args []sql.NamedArg) {
	return tpl.AnalyzeTPLFromCache(ctx, template, input, ctx.Placeholder())
}

func (ctx *MssqlContext) Placeholder() tpl.Placeholder {
	return &mssqlPlaceHolder{ctx: ctx}
}

func (ctx *MssqlContext) AnalyzeTPL(template string, input map[string]interface{}, ph tpl.Placeholder) (string, *tpl.ReplaceItem) {
	return tpl.DefaultAnalyze(ctx.symbols, template, input, ph)
}

func (ctx *MssqlContext) HandleAndSymbols(template string, rpsitem *tpl.ReplaceItem, input map[string]interface{}) (sql string, values []sql.NamedArg, exists bool) {
	word := regexp.MustCompile(tpl.AndPattern)
	item := rpsitem.Clone()
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

func (ctx *MssqlContext) HandleOrSymbols(template string, rpsitem *tpl.ReplaceItem, input map[string]interface{}) (sql string, values []sql.NamedArg, exists bool) {
	word := regexp.MustCompile(tpl.OrPattern)
	item := rpsitem.Clone()
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
