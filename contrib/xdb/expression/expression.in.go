package expression

import (
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &inExpressionMatcher{}

func NewInExpressionMatcher(symbolMap xdb.SymbolMap) xdb.ExpressionMatcher {
	//in aaa
	//in t.aaa
	//t.aaa in aaa
	//bbb in aaa
	const pattern = `[&|\|](({in\s+(\w+(\.\w+)?)})|({\w+(\.\w+)?)\s+in\s+(\w+)})`
	return &inExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
	}
}

type inExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
}

func (m *inExpressionMatcher) Name() string {
	return "in"
}

func (m *inExpressionMatcher) Pattern() string {
	return m.regexp.String()
}
func (m *inExpressionMatcher) Symbol() xdb.SymbolMap {
	return m.symbolMap
}
func (m *inExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {

	parties := m.regexp.FindStringSubmatch(expression)
	if len(parties) <= 0 {
		return
	}
	ok = true
	var (
		item = &xdb.ExpressionItem{
			Oper:   m.Name(),
			Symbol: getExpressionSymbol(expression),
		}
	)
	fullField := parties[2]

	fullField = strings.TrimSpace(fullField)
	item.FullField = fullField
	item.PropName = getExpressionPropertyName(fullField)

	item.ExpressionBuildCallback = m.buildCallback()
	return item, ok
}

func (m *inExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(item *xdb.ExpressionItem, param xdb.DBParam, argName string) (expression string, err xdb.MissError) {

		return
	}
}
