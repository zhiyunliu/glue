package tpl

import "database/sql"

type cacheItem struct {
	sql           string
	names         []string
	nameCache     map[string]string
	hasReplace    bool
	hasDynamicAnd bool
	hasDynamicOr  bool
	ph            Placeholder
	SQLTemplate   SQLTemplate
}

type ReplaceItem struct {
	Names       []string
	Values      []sql.NamedArg
	NameCache   map[string]string
	Placeholder Placeholder
}

func (p *ReplaceItem) Clone() *ReplaceItem {
	return &ReplaceItem{
		NameCache:   p.NameCache,
		Placeholder: p.Placeholder,
	}
}

func (item cacheItem) ClonePlaceHolder() Placeholder {
	return item.ph.Clone()
}

func (item cacheItem) build(input DBParam) (execSql string, values []sql.NamedArg) {
	values = make([]sql.NamedArg, len(item.names))
	ph := item.ClonePlaceHolder()
	for i := range item.names {
		_, values[i] = input.Get(item.names[i], ph)
	}

	rspitem := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ph,
	}
	for k := range item.nameCache {
		rspitem.NameCache[k] = item.nameCache[k]
	}

	execSql = item.sql
	if item.hasReplace {
		execSql, _ = handleRelaceSymbols(item.sql, input, ph)
	}
	var vals []sql.NamedArg
	if item.hasDynamicAnd {
		execSql, vals, _ = item.SQLTemplate.HandleAndSymbols(execSql, rspitem, input)
		values = append(values, vals...)
	}
	if item.hasDynamicOr {
		execSql, vals, _ = item.SQLTemplate.HandleOrSymbols(execSql, rspitem, input)
		values = append(values, vals...)
	}
	return execSql, values
}
