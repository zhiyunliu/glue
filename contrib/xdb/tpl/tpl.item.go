package tpl

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

func (item cacheItem) build(input DBParam) (execSql string, values []interface{}) {
	values = make([]interface{}, len(item.names))
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

	return execSql, values
}
