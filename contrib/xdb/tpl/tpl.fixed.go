package tpl

// FixedContext  模板
type FixedContext struct {
	name    string
	prefix  string
	symbols SymbolMap
}

type fixedPlaceHolder struct {
	ctx *FixedContext
}

func (ph *fixedPlaceHolder) Get(propName string) (argName, phName string) {
	phName = ph.ctx.prefix
	argName = propName
	return
}
func (ph *fixedPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	return val
}

func (ph *fixedPlaceHolder) NamedArg(propName string) (phName string) {
	phName = ph.ctx.prefix
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
		symbols: defaultSymbols.Clone(),
	}
}

func (ctx FixedContext) Name() string {
	return ctx.name
}

// GetSQLContext 获取查询串
func (ctx *FixedContext) GetSQLContext(tpl string, input map[string]interface{}) (query string, args []any, err error) {
	return AnalyzeTPLFromCache(ctx, tpl, input, ctx.Placeholder())
}

func (ctx *FixedContext) Placeholder() Placeholder {
	return &fixedPlaceHolder{ctx: ctx}
}

func (ctx *FixedContext) AnalyzeTPL(tpl string, input map[string]interface{}, ph Placeholder) (sql string, item *ReplaceItem, err error) {
	return DefaultAnalyze(ctx.symbols, tpl, input, ph)
}

func (ctx *FixedContext) RegisterSymbol(symbol Symbol) error {
	return ctx.symbols.Register(symbol)
}
