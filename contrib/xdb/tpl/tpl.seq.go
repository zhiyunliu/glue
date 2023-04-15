package tpl

import (
	"database/sql"
	"fmt"
	"regexp"
)

// SeqContext 参数化时使用@+参数名作为占位符的SQL数据库如:oracle,sql server
type SeqContext struct {
	name    string
	prefix  string
	symbols Symbols
}

type seqPlaceHolder struct {
	ctx *SeqContext
	idx int
}

func (ph *seqPlaceHolder) Get(propName string) (argName, phName string) {
	ph.idx++
	phName = fmt.Sprint(ph.ctx.prefix, ph.idx)
	argName = "p_" + propName
	return
}

func (ph *seqPlaceHolder) Clone() Placeholder {
	return &seqPlaceHolder{
		idx: ph.idx,
		ctx: ph.ctx,
	}
}

func NewSeq(name, prefix string) SQLTemplate {
	return &SeqContext{
		name:    name,
		prefix:  prefix,
		symbols: defaultSymbols,
	}
}

func (ctx SeqContext) Name() string {
	return ctx.name
}

// GetSQLContext 获取查询串
func (ctx *SeqContext) GetSQLContext(tpl string, input map[string]interface{}) (sql string, args []sql.NamedArg) {
	return AnalyzeTPLFromCache(ctx, tpl, input, ctx.Placeholder())
}

func (ctx *SeqContext) Placeholder() Placeholder {
	return &seqPlaceHolder{ctx: ctx, idx: 0}
}

func (ctx *SeqContext) AnalyzeTPL(tpl string, input map[string]interface{}, ph Placeholder) (sql string, item *ReplaceItem) {
	return DefaultAnalyze(ctx.symbols, tpl, input, ph)
}

func (ctx *SeqContext) HandleAndSymbols(template string, rspitem *ReplaceItem, input map[string]interface{}) (sql string, values []sql.NamedArg, exists bool) {
	word, _ := regexp.Compile(AndPattern)
	item := rspitem.Clone()
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

func (ctx *SeqContext) HandleOrSymbols(template string, rspitem *ReplaceItem, input map[string]interface{}) (sql string, values []sql.NamedArg, exists bool) {
	word := regexp.MustCompile(OrPattern)
	item := rspitem.Clone()
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
