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
	const pattern = `[&|\|]({\w+(\.\w+)?})`
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

func (m *normalExpressionMatcher) Symbol() xdb.SymbolMap {
	return m.symbolMap
}

func (m *normalExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {
	fullkey := expression
	ok = m.regexp.MatchString(fullkey)
	if !ok {
		return
	}

	parties := strings.Split(fullkey, ".")

	item := &xdb.ExpressionItem{
		Oper:      "=",
		FullField: fullkey,
		PropName:  fullkey,
	}
	if len(parties) > 1 {
		item.PropName = parties[1]
	}

	item.ExpressionBuildCallback = m.buildCallback()
	return item, ok
}

func (m *normalExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(item *xdb.ExpressionItem, param xdb.DBParam, argName string) (expression string, err xdb.MissError) {
		return
	}
}
