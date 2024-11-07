package expression

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &normalExpressionMatcher{}

func NewNormalExpressionMatcher(symbolMap xdb.SymbolMap) xdb.ExpressionMatcher {
	const pattern = `[$|@|&|\|]({(\w+(\.\w+)?\s*)})`
	matcher := &normalExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
	}
	matcher.operatorMap = matcher.getOperatorMap()
	return matcher
}

type normalExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	operatorMap     xdb.OperatorMap
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
	item.SpecConcat(m.symbolMap)
	item.ExpressionBuildCallback = m.buildCallback()
	m.expressionCache.Store(expression, item)

	return item, ok
}

func (m *normalExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(state xdb.SqlState, item *xdb.ExpressionItem, param xdb.DBParam, argName string, value any) (expression string, err xdb.MissError) {
		if !strings.EqualFold(item.GetSymbol(), xdb.SymbolReplace) {
			state.AppendExpr(argName, value)
		}

		callback, ok := m.operatorMap.Load(item.Symbol)
		if !ok {
			err = xdb.NewMissOperError(item.Oper)
			return
		}
		return callback(item, param, argName, value), nil
	}
}

func (m *normalExpressionMatcher) getOperatorMap() xdb.OperatorMap {
	var normalOperMap = xdb.NewOperatorMap()

	normalOperMap.Store("@", func(item xdb.ExpressionValuer, param xdb.DBParam, argName string, value any) string {
		return argName
	})

	normalOperMap.Store("&", func(item xdb.ExpressionValuer, param xdb.DBParam, argName string, value any) string {
		return fmt.Sprintf("%s %s=%s", item.GetConcat(), item.GetFullfield(), argName)
	})

	normalOperMap.Store("|", func(item xdb.ExpressionValuer, param xdb.DBParam, argName string, value any) string {
		return fmt.Sprintf("%s %s=%s", item.GetConcat(), item.GetFullfield(), argName)
	})

	normalOperMap.Store("$", func(item xdb.ExpressionValuer, param xdb.DBParam, argName string, value any) (val string) {

		switch t := value.(type) {
		case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
			val = strings.Trim(strings.Replace(fmt.Sprint(t), " ", ",", -1), "[]")
		case []string:
			val = sqlInjectionPreventionArray(t)
		default:
			val = fmt.Sprintf("%v", t)
			val = sqlInjectionPrevention(val)
		}
		return val
	})

	return normalOperMap
}
