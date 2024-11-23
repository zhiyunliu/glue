package expression

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &likeExpressionMatcher{}

func NewLikeExpressionMatcher(symbolMap xdb.SymbolMap, opts ...xdb.MatcherOption) xdb.ExpressionMatcher {
	//aaaa like ttt
	//aaaa like %ttt
	//aaaa like ttt%
	//aaaa like %ttt%
	//tt.aaaa like bbb
	//tt.aaaa like %bbb
	//tt.aaaa like bbb%
	//tt.aaaa like %bbb%

	//like ttt
	//like %ttt
	//like ttt%
	//like %ttt%
	//like t.bbb
	//like %t.bbb
	//like t.bbb%
	//like %t.bbb%

	mopts := &xdb.MatcherOptions{}
	for i := range opts {
		opts[i](mopts)
	}

	const pattern = `[&|\|](({(like|notlike)\s+(%?\w+(\.\w+)?%?)})|({(\w+(\.\w+)?)\s+(like|notlike)\s+(%?\w+%?)}))`

	matcher := &likeExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
		buildCallback:   mopts.BuildCallback,
	}
	matcher.operatorMap = matcher.getOperatorMap(mopts.OperatorMap)

	return matcher
}

type likeExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	operatorMap     xdb.OperatorMap
	buildCallback   xdb.ExpressionBuildCallback
}

func (m *likeExpressionMatcher) Name() string {
	return "like"
}

func (m *likeExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *likeExpressionMatcher) GetOperatorMap() xdb.OperatorMap {
	return m.operatorMap
}

func (m *likeExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {
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
	const SPEC_CHAR = "%"
	var (
		prefix   string
		suffix   string
		oper     string
		fullkey  string
		propName string
	)
	if parties[4] != "" {
		oper = parties[3]
		propName = parties[4]
		fullkey = strings.Trim(propName, SPEC_CHAR)
	} else {
		oper = parties[9]
		fullkey = parties[7]
		propName = parties[10]
	}

	if strings.HasPrefix(propName, SPEC_CHAR) {
		prefix = SPEC_CHAR
	}
	if strings.HasSuffix(propName, SPEC_CHAR) {
		suffix = SPEC_CHAR
	}

	oper = prefix + oper + suffix
	propName = strings.Trim(propName, SPEC_CHAR)

	item := &xdb.ExpressionItem{
		Symbol:    getExpressionSymbol(m.symbolMap, expression),
		Matcher:   m,
		Oper:      oper,
		FullField: fullkey,
		PropName:  getExpressionPropertyName(propName),
	}
	item.ExpressionBuildCallback = m.defaultBuildCallback()
	if m.buildCallback != nil {
		item.ExpressionBuildCallback = m.buildCallback
	}
	m.expressionCache.Store(expression, item)

	return item, ok
}

func (m *likeExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
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

func (m *likeExpressionMatcher) getOperatorMap(optMap xdb.OperatorMap) xdb.OperatorMap {

	operList := []xdb.Operator{
		xdb.NewOperator("like", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s like %s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("%like", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s like '%%'+%s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("like%", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s like %s+'%%'", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("%like%", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s like '%%'+%s+'%%'", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("notlike", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s not like %s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("%notlike", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s not like '%%'+%s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("notlike%", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s not like %s+'%%'", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("%notlike%", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s not like '%%'+%s+'%%'", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),
	}

	if optMap != nil {
		optMap.Range(func(name string, operator xdb.Operator) bool {
			operList = append(operList, operator)
			return true
		})
	}
	return xdb.NewOperatorMap(operList...)

}
