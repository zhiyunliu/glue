package sqlserver

import (
	"database/sql"
	"fmt"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
)

// MssqlContext  模板
type MssqlContext struct {
	name    string
	prefix  string
	symbols tpl.SymbolMap
}

type mssqlPlaceHolder struct {
	ctx *MssqlContext
}

func (ph *mssqlPlaceHolder) Get(propName string) (argName, phName string) {
	argName = fmt.Sprint(ph.ctx.prefix, propName)
	phName = "@" + argName
	return
}

func (ph *mssqlPlaceHolder) NamedArg(argName string) (phName string) {
	phName = "@" + argName
	return
}

func (ph *mssqlPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	if arg, ok := val.(sql.NamedArg); ok {
		return arg
	}
	return sql.NamedArg{Name: argName, Value: val}

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
func (ctx *MssqlContext) GetSQLContext(template string, input map[string]interface{}) (query string, args []any, err error) {
	return tpl.AnalyzeTPLFromCache(ctx, template, input, ctx.Placeholder())
}

func (ctx *MssqlContext) Placeholder() tpl.Placeholder {
	return &mssqlPlaceHolder{ctx: ctx}
}

func (ctx *MssqlContext) AnalyzeTPL(template string, input map[string]interface{}, ph tpl.Placeholder) (string, *tpl.ReplaceItem, error) {
	return tpl.DefaultAnalyze(ctx.symbols, template, input, ph)
}

func (ctx *MssqlContext) RegisterSymbol(symbol tpl.Symbol) error {
	return ctx.symbols.Register(symbol)
}
