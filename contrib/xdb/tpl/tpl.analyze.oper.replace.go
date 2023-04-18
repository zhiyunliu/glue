package tpl

import (
	"fmt"
	"regexp"
)

// 处理替换符合
func handleRelaceSymbols(tpl string, input map[string]interface{}, ph Placeholder) (string, bool) {
	word := regexp.MustCompile(ReplacePattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ph,
	}
	hasReplace := false
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		hasReplace = true
		fullKey := s[2 : len(s)-1]
		return replaceSymbols(input, fullKey, item)
	})

	return sql, hasReplace
}

func replaceSymbols(input DBParam, fullKey string, item *ReplaceItem) string {
	propName := GetPropName(fullKey)
	value := input.GetVal(propName)
	if !IsNil(value) {
		return fmt.Sprintf("%v", value)
	}
	return ""
}
