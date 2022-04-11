package tpl

//FixedContext  模板
type FixedContext struct {
	name   string
	prefix string
}

func (ctx FixedContext) Name() string {
	return ctx.name
}

//GetSQLContext 获取查询串
func (ctx *FixedContext) GetSQLContext(tpl string, input map[string]interface{}) (query string, args []interface{}) {
	return AnalyzeTPLFromCache(ctx, tpl, input)
}

func (ctx *FixedContext) Placeholder() Placeholder {
	return func() string { return ctx.prefix }
}

func (ctx *FixedContext) analyzeTPL(tpl string, input map[string]interface{}) (sql string, params []interface{}, names []string) {
	return defaultAnalyze(tpl, input, ctx.Placeholder())
}
