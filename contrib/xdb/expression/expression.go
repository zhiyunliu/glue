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

func sqlInjectionPrevention(data string) (newdata string) {
	newdata = strings.ReplaceAll(data, "'", "''")
	return
}

func sqlInjectionPreventionArray(data []string) (newdata string) {

	newArray := make([]string, len(data))
	for i := range data {
		newArray[i] = strings.ReplaceAll(data[i], "'", "''")
	}
	return "'" + strings.Join(newArray, "','") + "'"
}
