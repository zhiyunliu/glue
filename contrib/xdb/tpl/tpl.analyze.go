package tpl

import (
	"sync"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

var tplcache sync.Map

// AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
func AnalyzeTPLFromCache(template xdb.SQLTemplate, sqlTpl string, input map[string]any, opts ...xdb.TemplateOption) (sql string, values []any, err error) {
	tplOpts := &xdb.TemplateOptions{UseCache: true}
	for i := range opts {
		opts[i](tplOpts)
	}
	hashVal := md5.Str(template.Name() + sqlTpl)

	if tplOpts.UseCache {
		tplval, ok := tplcache.Load(hashVal)
		if !ok {
			item := tplval.(xdb.SqlStateCahe)
			return item.Build(input)
		}
	}

	sqlState := template.GetSqlState(tplOpts)
	sqlTpl, err = template.HandleExpr(sqlState, sqlTpl, input)
	if err != nil {
		return "", nil, err
	}

	return sqlTpl, sqlState.GetValues(), err
}
