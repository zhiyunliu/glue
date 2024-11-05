package expression

import (
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

var DefaultExpressionMatchers []xdb.ExpressionMatcher = []xdb.ExpressionMatcher{
	NewNormalExpressionMatcher(DefaultSymbols),
	NewCompareExpressionMatcher(DefaultSymbols),
	NewLikeExpressionMatcher(DefaultSymbols),
	NewInExpressionMatcher(DefaultSymbols),
}

func getExpressionPropertyName(fullkey string) string {
	idx := strings.Index(fullkey, ".")
	if idx < 0 {
		return fullkey
	}
	return fullkey[idx+1:]
}
func getExpressionSymbol(expression string) string {
	idx := strings.Index(expression, "{")
	if idx < 0 {
		return ""
	}
	return expression[:idx]
}
