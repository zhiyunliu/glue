package tpl

import (
	"fmt"
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

func (ph *seqPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	return val
}

func (ph *seqPlaceHolder) NamedArg(propName string) (phName string) {
	ph.idx++
	phName = fmt.Sprint(ph.ctx.prefix, ph.idx)
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
func (ctx *SeqContext) GetSQLContext(tpl string, input map[string]interface{}) (sql string, args []any) {
	return AnalyzeTPLFromCache(ctx, tpl, input, ctx.Placeholder())
}

func (ctx *SeqContext) Placeholder() Placeholder {
	return &seqPlaceHolder{ctx: ctx, idx: 0}
}

func (ctx *SeqContext) AnalyzeTPL(tpl string, input map[string]interface{}, ph Placeholder) (sql string, item *ReplaceItem) {
	return DefaultAnalyze(ctx.symbols, tpl, input, ph)
}
