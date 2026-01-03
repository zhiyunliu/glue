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
	pattern := ComparePattern
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
		Symbol:  GetExpressionSymbol(m.symbolMap, expression),
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
		item.PropName = GetExpressionPropertyName(item.FullField)
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
			//没有值，并且是可空
			if item.GetSymbol().IsDynamic() {
				return "", nil
			}
			return
		}
		err = nil
		if xdb.CheckIsNil(value) && item.GetSymbol().IsDynamic() {
			return
		}
		normalizeCall, ok := item.GetOperValueNormalizeCallback()
		if !ok {
			err = xdb.NewMissOperError(item.GetOper())
			return
		}
		value, err = normalizeCall(item, param, value)
		if err != nil {
			return
		}

		phName := state.AppendExpr(item, value)

		operCallback, ok := item.GetOperExprCallback()
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
		xdb.NewOperator(">", operCallback, nil),
		xdb.NewOperator(">=", operCallback, nil),
		xdb.NewOperator("<>", operCallback, nil),
		xdb.NewOperator("!=", operCallback, nil),
		xdb.NewOperator("=", operCallback, nil),
		xdb.NewOperator("<", operCallback, nil),
		xdb.NewOperator("<=", operCallback, nil),
	}

	if optMap != nil {
		optMap.Range(func(name string, operator xdb.Operator) bool {
			operList = append(operList, operator)
			return true
		})
	}
	return xdb.NewOperatorMap(operList...)
}
