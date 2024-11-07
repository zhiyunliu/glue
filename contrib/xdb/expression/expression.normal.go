package expression

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	var normalOperMap = NewOperatorMap()

	normalOperMap.Store("@", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return argName
	})

	normalOperMap.Store("&", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s=%s", concat, item.GetFullfield(), argName)
	})

	normalOperMap.Store("|", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s=%s", concat, item.GetFullfield(), argName)
	})

}

var _ xdb.ExpressionMatcher = &normalExpressionMatcher{}

func NewNormalExpressionMatcher(symbolMap xdb.SymbolMap) xdb.ExpressionMatcher {
	const pattern = `[$|@|&|\|]({(\w+(\.\w+)?\s*)})`
	return &normalExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
	}
}

type normalExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
}

func (m *normalExpressionMatcher) Name() string {
	return "normal"
}

func (m *normalExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *normalExpressionMatcher) LoadSymbol(symbol string) (xdb.Symbol, bool) {
	return m.symbolMap.Load(symbol)
}

func (m *normalExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {
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

	fullkey := strings.TrimSpace(parties[2])

	item := &xdb.ExpressionItem{
		Symbol:    getExpressionSymbol(expression),
		Oper:      "=",
		FullField: fullkey,
		PropName:  fullkey,
	}
	pIdx := strings.Index(fullkey, ".")

	if pIdx > 0 {
		item.PropName = fullkey[pIdx+1:]
	}

	item.ExpressionBuildCallback = m.buildCallback()
	m.expressionCache.Store(expression, item)

	return item, ok
}

func (m *normalExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(item *xdb.ExpressionItem, param xdb.DBParam, argName string) (expression string, err xdb.MissError) {
		return
	}
}
