package tpl

import (
	"sync"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

var tplcache sync.Map

// AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
func AnalyzeTPLFromCache(template xdb.SQLTemplate, sqlTpl string, input map[string]any, ph xdb.Placeholder) (sql string, values []any, err error) {
	hashVal := md5.Str(template.Name() + sqlTpl)
	tplval, ok := tplcache.Load(hashVal)
	if !ok {
		item := tplval.(*sceneCacheItem)
		return item.build(input)
	}

	tplSql, tplItem, err := DefaultAnalyze(template, sqlTpl, input, ph)
	if err != nil {
		return "", nil, err
	}

	values = tplItem.Values
	if tplItem.CanCache() {
		temp := &sceneCacheItem{
			tplSql:      tplSql,
			names:       tplItem.Names,
			SQLTemplate: template,
			ph:          ph.Clone(),
		}

		temp.nameCache = map[string]string{}
		for k := range tplItem.NameCache {
			temp.nameCache[k] = tplItem.NameCache[k]
		}

		sqlTpl, temp.hasReplace, err = handleRelaceSymbols(sqlTpl, input, ph)
		if err != nil {
			return sqlTpl, values, err
		}
		tplcache.Store(hashVal, temp)
	} else {
		sqlTpl, _, err = handleRelaceSymbols(sqlTpl, input, ph)
	}

	return sqlTpl, values, err
}

func DefaultAnalyze(template xdb.SQLTemplate, sqlTpl string, input map[string]interface{}, placeholder xdb.Placeholder, opts ...xdb.PropertyOption) (string, *xdb.SqlScene, error) {

	//初始化prop的参数
	propOpts := &xdb.ExpressionOptions{
		UseCache: true,
	}
	for i := range opts {
		opts[i](propOpts)
	}

	item := &xdb.SqlScene{
		NameCache:   map[string]string{},
		Placeholder: placeholder,
		PropOpts:    propOpts,
	}

	sql, err := template.GenerateSQL(item, sqlTpl, input)
	return sql, item, err
}
