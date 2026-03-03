package expression

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &inExpressionMatcher{}

func NewInExpressionMatcher(symbolMap xdb.SymbolMap, opts ...xdb.MatcherOption) xdb.ExpressionMatcher {
	//in aaa
	//in t.aaa
	//t.aaa in aaa
	//bbb in aaa

	mopts := &xdb.MatcherOptions{}
	for i := range opts {
		opts[i](mopts)
	}
	pattern := InPattern
	matcher := &inExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
		buildCallback:   mopts.BuildCallback,
	}
	matcher.operatorMap = matcher.getOperatorMap(mopts.OperatorMap)

	return matcher
}

type inExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	buildCallback   xdb.ExpressionBuildCallback
	operatorMap     xdb.OperatorMap
}

func (m *inExpressionMatcher) Name() string {
	return "in"
}

func (m *inExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *inExpressionMatcher) GetOperatorMap() xdb.OperatorMap {
	return m.operatorMap
}

func (m *inExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {
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

	var (
		item = &xdb.ExpressionItem{
			Symbol:  GetExpressionSymbol(m.symbolMap, expression),
			Matcher: m,
		}
		fullField string
		propName  string
		oper      string
	)
	// fullfield,oper,oper
	//&{in tbl.field} => 3,in,prop(3)
	//&{tt.field  in    property} => 6,in, 8

	if parties[4] != "" {
		oper = parties[3]
		fullField = parties[4]
		propName = GetExpressionPropertyName(fullField)

	} else {
		oper = parties[9]
		fullField = parties[7]
		propName = parties[10]
	}

	item.FullField = fullField
	item.PropName = propName
	item.Oper = strings.ReplaceAll(oper, " ", "")

	item.ExpressionBuildCallback = m.defaultBuildCallback()
	if m.buildCallback != nil {
		item.ExpressionBuildCallback = m.buildCallback
	}

	m.expressionCache.Store(expression, item)
	return item, ok
}

func (m *inExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(item xdb.ExpressionValuer, state xdb.SqlState, param xdb.DBParam) (expression string, err xdb.MissError) {
		value, err := param.GetVal(item.GetPropName())
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

		operCallback, ok := item.GetOperExprCallback()
		if !ok {
			err = xdb.NewMissOperError(item.GetOper())
			return
		}
		return operCallback(item, param, "", value), nil
	}
}
func (m *inExpressionMatcher) getOperatorMap(optMap xdb.OperatorMap) xdb.OperatorMap {

	inCallback := func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s in (%s)", item.GetSymbol().Concat(), item.GetFullfield(), value)
	}
	notinCallback := func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s not in (%s)", item.GetSymbol().Concat(), item.GetFullfield(), value)
	}

	emptyNormalize := func(exprName xdb.ExprName, param xdb.DBParam, value any) (newVal any, err xdb.MissError) {
		var val string
		switch t := value.(type) {
		case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
			val = strings.Trim(strings.ReplaceAll(fmt.Sprint(t), " ", ","), "[]")
			if len(val) == 0 {
				return
			}
		case []string:
			if len(t) <= 0 {
				return
			}
			val = sqlInjectionPreventionArray(t)
		case []byte:
			return "", xdb.NewMissDataTypeError(exprName.GetPropName())
		default:
			refVal := reflect.ValueOf(value)
			if !(refVal.Kind() == reflect.Array ||
				refVal.Kind() == reflect.Slice) {
				return "", xdb.NewMissDataTypeError(exprName.GetPropName())
			}
			arrayLen := refVal.Len()
			if arrayLen <= 0 {
				return
			}
			tmpStrArray := make([]string, arrayLen)
			for i := 0; i < arrayLen; i++ {
				ele := refVal.Index(i)
				tmpStrArray[i] = fmt.Sprint(ele.Interface())
			}
			val = sqlInjectionPreventionArray(tmpStrArray)
		}
		return val, nil
	}

	operList := []xdb.Operator{
		xdb.NewOperator("in", inCallback, emptyNormalize),
		xdb.NewOperator("notin", notinCallback, emptyNormalize),
	}

	if optMap != nil {
		optMap.Range(func(name string, operator xdb.Operator) bool {
			operList = append(operList, operator)
			return true
		})
	}
	return xdb.NewOperatorMap(operList...)

}
