package expression

import (
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.ExpressionMatcher = &likeExpressionMatcher{}

func NewLikeExpressionMatcher(symbolMap xdb.SymbolMap) xdb.ExpressionMatcher {
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

	const pattern = `[&|\|](({like\s+(%?\w+(\.\w+)?%?)})|({(\w+(\.\w+)?)\s+like\s+(%?\w+%?)}))`
	return &likeExpressionMatcher{
		regexp:          regexp.MustCompile(pattern),
		expressionCache: &sync.Map{},
		symbolMap:       symbolMap,
	}
}

type likeExpressionMatcher struct {
	symbolMap       xdb.SymbolMap
	regexp          *regexp.Regexp
	expressionCache *sync.Map
}

func (m *likeExpressionMatcher) Name() string {
	return "like"
}

func (m *likeExpressionMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *likeExpressionMatcher) LoadSymbol(symbol string) (xdb.Symbol, bool) {
	return m.symbolMap.Load(symbol)
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
		oper     string = m.Name()
		fullkey  string
		propName string
	)
	if parties[3] != "" {

		propName = parties[3]
		fullkey = strings.Trim(propName, SPEC_CHAR)
	} else {
		fullkey = parties[6]
		propName = parties[8]
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
		Symbol:    getExpressionSymbol(expression),
		Oper:      oper,
		FullField: fullkey,
		PropName:  getExpressionPropertyName(propName),
	}
	item.ExpressionBuildCallback = m.buildCallback()
	m.expressionCache.Store(expression, item)

	return item, ok
}

func (m *likeExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(item *xdb.ExpressionItem, param xdb.DBParam, argName string) (expression string, err xdb.MissError) {
		return
	}
}
