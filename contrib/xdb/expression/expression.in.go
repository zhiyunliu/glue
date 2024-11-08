package expression

import (
	"fmt"
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

	const pattern = `[&|\|](({in\s+(\w+(\.\w+)?)\s*})|({(\w+(\.\w+)?)\s+in\s+(\w+)\s*}))`
	return &inExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
		buildCallback:   mopts.BuildCallback,
	}
}

type inExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
	buildCallback   xdb.ExpressionBuildCallback
}

func (m *inExpressionMatcher) Name() string {
	return "in"
}

func (m *inExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *inExpressionMatcher) LoadSymbol(symbol string) (xdb.Symbol, bool) {
	return m.symbolMap.Load(symbol)
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
			Symbol: getExpressionSymbol(expression),
			Oper:   m.Name(),
		}
		fullField string
		propName  string
	)
	// fullfield,oper,oper
	//&{in tbl.field} => 3,in,prop(3)
	//&{tt.field  in    property} => 6,in, 8

	if parties[3] != "" {
		fullField = parties[3]
		propName = getExpressionPropertyName(fullField)
	} else {
		fullField = parties[6]
		propName = parties[8]
	}

	item.FullField = fullField
	item.PropName = propName

	item.SpecConcat(m.symbolMap)
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
		default:
			return
		}

		return fmt.Sprintf("%s %s in (%s)", item.GetConcat(), item.GetFullfield(), val), nil
	}
}
