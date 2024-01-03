package tpl

import "github.com/zhiyunliu/glue/xdb"

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
	Values      []interface{}
	NameCache   map[string]string
	Placeholder Placeholder
	HasAndOper  bool
	HasOrOper   bool
}

func (p *ReplaceItem) Clone() *ReplaceItem {
	return &ReplaceItem{
		NameCache:   p.NameCache,
		Placeholder: p.Placeholder,
	}
}

func (p *ReplaceItem) CanCache() bool {
	return !(p.HasAndOper || p.HasOrOper)
}

func (item cacheItem) ClonePlaceHolder() Placeholder {
	return item.ph.Clone()
}

func (item cacheItem) build(input DBParam) (execSql string, values []interface{}, err error) {
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
		return "", values, xdb.NewMissParamsError(outerrs...)
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
		execSql, _, err = handleRelaceSymbols(item.sql, input, ph)
	}

	return execSql, values, err
}
