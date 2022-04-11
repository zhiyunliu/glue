package tpl

import (
	"fmt"
)

//SeqContext 参数化时使用@+参数名作为占位符的SQL数据库如:oracle,sql server
type SeqContext struct {
	name   string
	prefix string
}

func (ctx SeqContext) Name() string {
	return ctx.name
}

//GetSQLContext 获取查询串
func (ctx *SeqContext) GetSQLContext(tpl string, input map[string]interface{}) (sql string, args []interface{}) {
	return AnalyzeTPLFromCache(ctx, tpl, input)
}

func (ctx *SeqContext) Placeholder() Placeholder {
	index := 0
	f := func() string {
		index++
		return fmt.Sprint(ctx.prefix, index)
	}
	return f
}

func (ctx *SeqContext) analyzeTPL(tpl string, input map[string]interface{}) (sql string, params []interface{}, names []string) {
	return defaultAnalyze(tpl, input, ctx.Placeholder())
}
