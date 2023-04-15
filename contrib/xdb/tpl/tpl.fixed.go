package tpl

import (
	"database/sql"
	"regexp"
)

// FixedContext  模板
type FixedContext struct {
	name    string
	prefix  string
	symbols Symbols
}

type fixedPlaceHolder struct {
	ctx *FixedContext
}

func (ph *fixedPlaceHolder) Get(propName string) (argName, phName string) {
	phName = ph.ctx.prefix
	argName = propName
	return
}

func (ph *fixedPlaceHolder) Clone() Placeholder {
	return &fixedPlaceHolder{
		ctx: ph.ctx,
	}
}

func NewFixed(name, prefix string) SQLTemplate {
	return &FixedContext{
		name:    name,
		prefix:  prefix,
		symbols: defaultSymbols,
	}
}

func (ctx FixedContext) Name() string {
	return ctx.name
}

// GetSQLContext 获取查询串
func (ctx *FixedContext) GetSQLContext(tpl string, input map[string]interface{}) (query string, args []sql.NamedArg) {
	return AnalyzeTPLFromCache(ctx, tpl, input, ctx.Placeholder())
}

func (ctx *FixedContext) Placeholder() Placeholder {
	return &fixedPlaceHolder{ctx: ctx}
}

func (ctx *FixedContext) AnalyzeTPL(tpl string, input map[string]interface{}, ph Placeholder) (sql string, item *ReplaceItem) {
	return DefaultAnalyze(ctx.symbols, tpl, input, ph)
}

func (ctx *FixedContext) HandleAndSymbols(template string, rpsitem *ReplaceItem, input map[string]interface{}) (sql string, values []sql.NamedArg, exists bool) {
	word, _ := regexp.Compile(AndPattern)
	item := rpsitem.Clone()
	symbols := ctx.symbols
	exists = false
	//变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(template, func(s string) string {
		symbol := s[:1]
		fullKey := s[2 : len(s)-1]
		callback, ok := symbols[symbol]
		if !ok {
			return s
		}
		return callback(input, fullKey, item)
	})

	return sql, item.Values, len(item.Values) > 0
}

func (ctx *FixedContext) HandleOrSymbols(template string, rpsitem *ReplaceItem, input map[string]interface{}) (sql string, values []sql.NamedArg, exists bool) {
	word, _ := regexp.Compile(OrPattern)
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
