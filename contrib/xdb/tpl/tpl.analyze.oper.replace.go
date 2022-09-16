package tpl

import (
	"fmt"
	"regexp"
)

//处理替换符合
func handleRelaceSymbols(tpl string, input map[string]interface{}) (string, bool) {
	word, _ := regexp.Compile(`\$\{\w+[\.]?\w+\}`)
	item := &ReplaceItem{
		NameCache: map[string]string{},
	}
	hasReplace := false
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		hasReplace = true
		fullKey := s[2 : len(s)-1]
		return replaceSymbols(input, fullKey, item)
	})

	return sql, hasReplace
}

func replaceSymbols(input map[string]interface{}, fullKey string, item *ReplaceItem) string {
	propName := GetPropName(fullKey)
	value := input[propName]
	if !IsNil(value) {
		return fmt.Sprintf("%v", value)
	}
	return ""
}
