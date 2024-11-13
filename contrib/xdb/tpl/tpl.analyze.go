package tpl

import (
	"github.com/zhiyunliu/glue/xdb"
)

// AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
func AnalyzeTPLFromCache(template xdb.SQLTemplate, sqlTpl string, input map[string]any, opts ...xdb.TemplateOption) (sql string, values []any, err error) {
	tplOpts := &xdb.TemplateOptions{UseExprCache: true}
	for i := range opts {
		opts[i](tplOpts)
	}
	sqlState := template.GetSqlState(tplOpts)
	defer template.ReleaseSqlState(sqlState)
	sqlTpl, err = template.HandleExpr(sqlState, sqlTpl, input)
	if err != nil {
		return "", nil, err
	}

	return sqlTpl, sqlState.GetValues(), err
}
