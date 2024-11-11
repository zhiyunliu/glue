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

	const pattern = `[&|\|](({((\w+\.)?\w+)\s*(>|>=|<>|=|<|<=)\s*(\w+)})|({(>|>=|<>|=|<|<=)\s*(\w+(\.\w+)?)}))`

	matcher := &compareExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
		buildCallback:   mopts.BuildCallback,
	}

	matcher.operatorMap = matcher.getOperatorMap(mopts.OperatorMap)

	return matcher
}

type compareExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	buildCallback   xdb.ExpressionBuildCallback
	operatorMap     xdb.OperatorMap
}

func (m *compareExpressionMatcher) Name() string {
	return "compare"
}

func (m *compareExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *compareExpressionMatcher) GetOperatorMap() xdb.OperatorMap {
	return m.operatorMap
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
	item := &xdb.ExpressionItem{
		Symbol:  getExpressionSymbol(m.symbolMap, expression),
		Matcher: m,
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

	item.ExpressionBuildCallback = m.defaultBuildCallback()
	if m.buildCallback != nil {
		item.ExpressionBuildCallback = m.buildCallback
	}
	m.expressionCache.Store(expression, item)
	return item, ok
}

func (m *compareExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(item xdb.ExpressionValuer, state xdb.SqlState, param xdb.DBParam) (expression string, err xdb.MissError) {
		propName := item.GetPropName()
		value, err := param.GetVal(propName)
		if err != nil {
			return
		}
		if xdb.CheckIsNil(value) {
			return
		}

		phName := state.AppendExpr(propName, value)

		operCallback, ok := item.GetOperatorCallback()
		if !ok {
			err = xdb.NewMissOperError(item.GetOper())
			return
		}
		return operCallback(item, param, phName, value), nil
	}
}

func (m *compareExpressionMatcher) getOperatorMap(optMap xdb.OperatorMap) xdb.OperatorMap {

	operCallback := func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s%s%s", item.GetSymbol().Concat(), item.GetFullfield(), item.GetOper(), phName)
	}
	operList := []xdb.Operator{
		xdb.NewDefaultOperator(">", operCallback),
		xdb.NewDefaultOperator(">=", operCallback),
		xdb.NewDefaultOperator("<>", operCallback),
		xdb.NewDefaultOperator("=", operCallback),
		xdb.NewDefaultOperator("<", operCallback),
		xdb.NewDefaultOperator("<=", operCallback),
	}

	if optMap != nil {
		optMap.Range(func(name string, operator xdb.Operator) bool {
			operList = append(operList, operator)
			return true
		})
	}
	return xdb.NewOperatorMap(operList...)
}
