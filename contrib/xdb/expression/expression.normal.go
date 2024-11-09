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
		nilNeedArgMap: map[string]bool{
			xdb.SymbolAt: true,
		},
	}
	matcher.operatorMap = matcher.getOperatorMap()
	if mopts.OperatorMap != nil {
		mopts.OperatorMap.Range(func(k string, v xdb.OperatorCallback) bool {
			matcher.operatorMap.Store(k, v)
			return true
		})
	}
	return matcher
}

type normalExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	operatorMap     xdb.OperatorMap
	buildCallback   xdb.ExpressionBuildCallback
	nilNeedArgMap   map[string]bool
}

func (m *normalExpressionMatcher) Name() string {
	return "normal"
}

func (m *normalExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

// func (m *normalExpressionMatcher) LoadSymbol(symbol string) (xdb.Symbol, bool) {
// 	return m.symbolMap.Load(symbol)
// }

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
	item.ExpressionBuildCallback = m.defaultBuildCallback()
	if m.buildCallback != nil {
		item.ExpressionBuildCallback = m.buildCallback
	}

	m.expressionCache.Store(expression, item)

	return item, ok
}

func (m *normalExpressionMatcher) IsNilNeedArg(symbol string) bool {
	return m.nilNeedArgMap[symbol]
}

func (m *normalExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(item xdb.ExpressionValuer, state xdb.SqlState, param xdb.DBParam) (expression string, err xdb.MissError) {

		var (
			phName       string
			argName      string
			isNilNeedArg bool = m.IsNilNeedArg(item.GetSymbol())
		)

		propName := item.GetPropName()
		value, err := param.GetVal(propName)
		if err != nil {
			return
		}

		isNil := xdb.CheckIsNil(value)

		//是空&&不需要参数，则退出
		if isNil && !isNilNeedArg {
			return
		}

		if !strings.EqualFold(item.GetSymbol(), xdb.SymbolReplace) {
			placeHolder := state.GetPlaceholder()
			argName, phName = placeHolder.Get(propName)
			value = placeHolder.BuildArgVal(argName, value)
			state.AppendExpr(propName, value)
		}

		operCallback, ok := m.operatorMap.Load(item.GetSymbol())
		if !ok {
			err = xdb.NewMissOperError(item.GetOper())
			return
		}
		return operCallback(item, param, phName, value), nil
	}
}

func (m *normalExpressionMatcher) getOperatorMap() xdb.OperatorMap {
	var normalOperMap = xdb.NewOperatorMap()

	normalOperMap.Store("@", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return phName
	})

	normalOperMap.Store("&", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s=%s", item.GetConcat(), item.GetFullfield(), phName)
	})

	normalOperMap.Store("|", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s=%s", item.GetConcat(), item.GetFullfield(), phName)
	})

	normalOperMap.Store("$", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) (val string) {

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
