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

	const pattern = `[&|\|](({(in|notin)\s+(\w+(\.\w+)?)\s*})|({(\w+(\.\w+)?)\s+(in|notin)\s+(\w+)\s*}))`

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
			Symbol:  getExpressionSymbol(m.symbolMap, expression),
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
		propName = getExpressionPropertyName(fullField)

	} else {
		oper = parties[9]
		fullField = parties[7]
		propName = parties[10]
	}

	item.FullField = fullField
	item.PropName = propName
	item.Oper = oper

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
			return
		}
		if xdb.CheckIsNil(value) {
			return
		}

		var val string
		switch t := value.(type) {
		case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
			val = strings.Trim(strings.Replace(fmt.Sprint(t), " ", ",", -1), "[]")
			if len(val) == 0 {
				return
			}
		case []string:
			if len(t) <= 0 {
				return
			}
			val = sqlInjectionPreventionArray(t)
		case []byte:
			return "", xdb.NewMissDataTypeError(item.GetPropName())
		default:
			refVal := reflect.ValueOf(value)
			if !(refVal.Kind() == reflect.Array ||
				refVal.Kind() == reflect.Slice) {
				return "", xdb.NewMissDataTypeError(item.GetPropName())
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

		operCallback, ok := item.GetOperatorCallback()
		if !ok {
			err = xdb.NewMissOperError(item.GetOper())
			return
		}
		return operCallback(item, param, "", val), nil
	}
}
func (m *inExpressionMatcher) getOperatorMap(optMap xdb.OperatorMap) xdb.OperatorMap {

	inCallback := func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s in (%s)", item.GetSymbol().Concat(), item.GetFullfield(), value)
	}
	notinCallback := func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s not in (%s)", item.GetSymbol().Concat(), item.GetFullfield(), value)
	}

	operList := []xdb.Operator{
		xdb.NewOperator("in", inCallback),
		xdb.NewOperator("notin", notinCallback),
	}

	if optMap != nil {
		optMap.Range(func(name string, operator xdb.Operator) bool {
			operList = append(operList, operator)
			return true
		})
	}
	return xdb.NewOperatorMap(operList...)

}
