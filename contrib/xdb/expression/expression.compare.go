package expression

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &compareExpressionMatcher{}

func NewCompareExpressionMatcher(symbolMap xdb.SymbolMap, opts ...xdb.MatcherOption) xdb.ExpressionMatcher {
	//t.field < aaa
	//t.field > aaa
	//t.field <= aaa
	//t.field >= aaa

	// field < aaa
	// field > aaa
	// field <= aaa
	// field >= aaa

	mopts := &xdb.MatcherOptions{}
	for i := range opts {
		opts[i](mopts)
	}

	const pattern = `[&|\|](({((\w+\.)?\w+)\s*(>|>=|=|<|<=)\s*(\w+)})|({(>|>=|=|<|<=)\s*(\w+(\.\w+)?)}))`
	return &compareExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
		buildCallback:   mopts.BuildCallback,
	}
}

type compareExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	buildCallback   xdb.ExpressionBuildCallback
}

func (m *compareExpressionMatcher) Name() string {
	return "compare"
}

func (m *compareExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *compareExpressionMatcher) LoadSymbol(symbol string) (xdb.Symbol, bool) {
	return m.symbolMap.Load(symbol)
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
	//fullfield,oper,property
	//{t.field=property} =3，5,6
	//{<property} =9,8, get(9)
	//

	item := &xdb.ExpressionItem{
		Symbol: getExpressionSymbol(expression),
	}

	if parties[5] != "" {
		item.FullField = parties[3]
		item.Oper = parties[5]
		item.PropName = parties[6]
	}
	if parties[8] != "" {
		item.FullField = parties[9]
		item.Oper = parties[8]
		item.PropName = getExpressionPropertyName(item.FullField)
	}

	item.SpecConcat(m.symbolMap)
	item.ExpressionBuildCallback = m.defaultBuildCallback()
	if m.buildCallback != nil {
		item.ExpressionBuildCallback = m.buildCallback
	}
	m.expressionCache.Store(expression, item)
	return item, ok
}

func (m *compareExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(state xdb.SqlState, item *xdb.ExpressionItem, param xdb.DBParam, argName string, value any) (expression string, err xdb.MissError) {
		state.AppendExpr(argName, value)
		return fmt.Sprintf("%s %s%s%s", item.GetConcat(), item.GetFullfield(), item.GetOper(), argName), nil
	}
}
