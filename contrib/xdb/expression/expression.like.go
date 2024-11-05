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

	const pattern = `[&|\|]({like\s+%?\w+(\.\w+)?%?}|{\w+(\.\w+)?\s+like\s+%?\w+%?})`
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

func (m *likeExpressionMatcher) Symbol() xdb.SymbolMap {
	return m.symbolMap
}

func (m *likeExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {

	fullkey := strings.TrimPrefix(expression, "like")
	fullkey = strings.TrimSpace(fullkey)

	var (
		prefix string
		suffix string
		oper   string = "like"
	)

	if strings.HasPrefix(fullkey, "%") {
		prefix = "%"
	}
	if strings.HasSuffix(fullkey, "%") {
		suffix = "%"
	}

	oper = prefix + oper + suffix
	fullkey = strings.Trim(fullkey, "%")

	item := &xdb.ExpressionItem{
		Oper:      oper,
		Symbol:    getExpressionSymbol(expression),
		FullField: fullkey,
	}
	item.PropName = getExpressionPropertyName(fullkey)
	item.ExpressionBuildCallback = m.buildCallback()
	return item, ok
}

func (m *likeExpressionMatcher) buildCallback() xdb.ExpressionBuildCallback {
	return func(item *xdb.ExpressionItem, param xdb.DBParam, argName string) (expression string, err xdb.MissError) {
		return
	}
}
