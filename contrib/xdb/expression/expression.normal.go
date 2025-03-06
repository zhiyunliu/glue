package expression

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &normalExpressionMatcher{}

func NewNormalExpressionMatcher(symbolMap xdb.SymbolMap, opts ...xdb.MatcherOption) xdb.ExpressionMatcher {
	mopts := &xdb.MatcherOptions{}
	for i := range opts {
		opts[i](mopts)
	}

	const pattern = `[$|@|&|\|]({(\w+(\.\w+)?\s*)})`
	matcher := &normalExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
		buildCallback:   mopts.BuildCallback,
	}
	matcher.operatorMap = matcher.getOperatorMap(mopts.OperatorMap)

	return matcher
}

type normalExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	operatorMap     xdb.OperatorMap
	buildCallback   xdb.ExpressionBuildCallback
}

func (m *normalExpressionMatcher) Name() string {
	return "normal"
}

func (m *normalExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *normalExpressionMatcher) GetOperatorMap() xdb.OperatorMap {
	return m.operatorMap
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
		Symbol:    getExpressionSymbol(m.symbolMap, expression),
		Matcher:   m,
		FullField: fullkey,
		PropName:  fullkey,
	}
	item.Oper = item.Symbol.Name()
	pIdx := strings.Index(fullkey, ".")

	if pIdx > 0 {
		item.PropName = fullkey[pIdx+1:]
	}
	item.ExpressionBuildCallback = m.defaultBuildCallback()
	if m.buildCallback != nil {
		item.ExpressionBuildCallback = m.buildCallback
	}

	m.expressionCache.Store(expression, item)

	return item, ok
}

func (m *normalExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(item xdb.ExpressionValuer, state xdb.SqlState, param xdb.DBParam) (expression string, err xdb.MissError) {
		var (
			phName   string
			propName = item.GetPropName()
		)
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

		if !strings.EqualFold(item.GetSymbol().Name(), xdb.SymbolReplace) {
			phName = state.AppendExpr(propName, value)
		}

		operCallback, ok := item.GetOperatorCallback()
		if !ok {
			err = xdb.NewMissOperError(item.GetOper())
			return
		}
		return operCallback(item, param, phName, value), nil
	}
}

func (m *normalExpressionMatcher) getOperatorMap(optMap xdb.OperatorMap) xdb.OperatorMap {
	operList := []xdb.Operator{

		xdb.NewOperator("@", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return phName
		}),

		xdb.NewOperator("&", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s=%s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("|", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s=%s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),

		xdb.NewOperator("$", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) (val string) {

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
