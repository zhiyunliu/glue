package expression

import (
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

var (
	DefaultExpressionMatchers []xdb.ExpressionMatcher
	ComparePattern            = `[@|&|\|](({((\w+\.)?\w+)\s*(>|>=|<>|!=|=|<|<=)\s*(\w+)})|({(>|>=|<>|!=|=|<|<=)\s*(\w+(\.\w+)?)}))`
	InPattern                 = `[@|&|\|](({(in|not\s*in)\s+(\w+(\.\w+)?)\s*})|({(\w+(\.\w+)?)\s+(in|not\s*in)\s+(\w+)\s*}))`
	LikePattern               = `[@|&|\|](({(like|not\s*like)\s+(%?\w+(\.\w+)?%?)})|({(\w+(\.\w+)?)\s+(like|not\s*like)\s+(%?\w+%?)}))`
	NormalPattern             = `[$|@|&|\|]({(\w+(\.\w+)?\s*)})`
)

func init() {
	initSqlState()
	initOperator()
	initSymbols()

	DefaultExpressionMatchers = []xdb.ExpressionMatcher{
		NewNormalExpressionMatcher(DefaultSymbols),
		NewCompareExpressionMatcher(DefaultSymbols),
		NewLikeExpressionMatcher(DefaultSymbols),
		NewInExpressionMatcher(DefaultSymbols),
	}
}

func GetExpressionPropertyName(fullkey string) string {
	idx := strings.Index(fullkey, ".")
	if idx < 0 {
		return fullkey
	}
	return fullkey[idx+1:]
}

// getExpressionSymbol 可能存在崩溃，在开发阶段即可暴露，无需关注
func GetExpressionSymbol(symbolMap xdb.SymbolMap, expression string) xdb.Symbol {
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
