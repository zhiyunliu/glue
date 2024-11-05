package tpl

import "github.com/zhiyunliu/glue/xdb"

type sceneCacheItem struct {
	tplSql        string
	names         []string
	nameCache     map[string]string
	hasReplace    bool
	hasDynamicAnd bool
	hasDynamicOr  bool
	ph            xdb.Placeholder
	SQLTemplate   xdb.SQLTemplate
}

func (item *sceneCacheItem) ClonePlaceHolder() xdb.Placeholder {
	return item.ph.Clone()
}

func (item *sceneCacheItem) build(input xdb.DBParam) (execSql string, values []interface{}, err error) {
	values = make([]interface{}, len(item.names))
	ph := item.ClonePlaceHolder()
	var outerrs []xdb.MissError
	var ierr xdb.MissError
	for i := range item.names {
		_, values[i], ierr = input.Get(item.names[i], ph)
		if ierr != nil {
			outerrs = append(outerrs, ierr)
		}
	}
	if len(outerrs) > 0 {
		return "", values, xdb.NewMissListError(outerrs...)
	}

	rspitem := &xdb.SqlScene{
		NameCache:   map[string]string{},
		Placeholder: ph,
	}
	for k := range item.nameCache {
		rspitem.NameCache[k] = item.nameCache[k]
	}

	execSql = item.tplSql
	if item.hasReplace {
		execSql, _, err = handleRelaceSymbols(execSql, input, ph)
	}

	return execSql, values, err
}
