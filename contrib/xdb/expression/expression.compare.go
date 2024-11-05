package expression

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &compareExpressionMatcher{}

func NewCompareExpressionMatcher(symbolMap xdb.SymbolMap) xdb.ExpressionMatcher {
	//t.field < aaa
	//t.field > aaa
	//t.field <= aaa
	//t.field >= aaa

	// field < aaa
	// field > aaa
	// field <= aaa
	// field >= aaa

	const pattern = `[&|\|]({(((\w+\.)?\w+)\s*(>|>=|=|<|<=)\s*(\w+)})|({(>|>=|=|<|<=)\s*(\w+(\.\w+)?)}))`
	return &compareExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
	}
}

type compareExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
}

func (m *compareExpressionMatcher) Name() string {
	return "compare"
}

func (m *compareExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *compareExpressionMatcher) Symbol() xdb.SymbolMap {
	return m.symbolMap
}

func (m *compareExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {
	tmp, ok := m.expressionCache.Load(expression)
	if ok {
		valuer = tmp.(xdb.ExpressionValuer)
		return
	}

	parties := m.regexp.FindStringSubmatch(expression)
	if len(parties) <= 0 {
		return
	}
	ok = true

	item := &xdb.ExpressionItem{
		FullField: parties[1],
		Oper:      parties[2],
		PropName:  parties[3],
		Symbol:    parties[3],
	}
	if len(parties) == 5 {
		item.Oper = parties[3]
		item.PropName = parties[4]
	}

	item.ExpressionBuildCallback = m.buildCallback()
	return item, ok
}

func (m *compareExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(item *xdb.ExpressionItem, param xdb.DBParam, argName string) (expression string, err xdb.MissError) {
		symbol, ok := m.symbolMap.Load(item.GetSymbol())
		if !ok {
			return "", xdb.NewMissPropError(item.GetPropName())
		}

		return fmt.Sprintf("%s %s%s%s", symbol.Concat(), item.GetFullfield(), item.GetOper(), argName), nil
	}
}
