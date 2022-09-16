package tpl

import "regexp"

//FixedContext  模板
type FixedContext struct {
	name    string
	prefix  string
	symbols Symbols
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

//GetSQLContext 获取查询串
func (ctx *FixedContext) GetSQLContext(tpl string, input map[string]interface{}) (query string, args []interface{}) {
	return AnalyzeTPLFromCache(ctx, tpl, input)
}

func (ctx *FixedContext) Placeholder() Placeholder {
	return func() string { return ctx.prefix }
}

func (ctx *FixedContext) AnalyzeTPL(tpl string, input map[string]interface{}) (sql string, names []string, values []interface{}) {
	return DefaultAnalyze(ctx.symbols, tpl, input, ctx.Placeholder())
}

func (ctx *FixedContext) HandleAndSymbols(template string, input map[string]interface{}) (sql string, values []interface{}, exists bool) {
	word, _ := regexp.Compile(AndPattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ctx.Placeholder(),
	}
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

func (ctx *FixedContext) HandleOrSymbols(template string, input map[string]interface{}) (sql string, values []interface{}, exists bool) {
	word, _ := regexp.Compile(OrPattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ctx.Placeholder(),
	}
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
