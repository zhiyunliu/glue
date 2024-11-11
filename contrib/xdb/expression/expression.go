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

// getExpressionSymbol 可能存在崩溃，在开发阶段即可暴露，无需关注
func getExpressionSymbol(symbolMap xdb.SymbolMap, expression string) xdb.Symbol {
	idx := strings.Index(expression, "{")
	if idx < 0 {
		return nil
	}
	symbol, _ := symbolMap.Load(expression[:idx])
	return symbol
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
