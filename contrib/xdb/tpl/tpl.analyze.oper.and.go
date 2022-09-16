package tpl

import (
	"fmt"
	"regexp"
)

//处理& 符号
func handleAndSymbols(tpl string, input map[string]interface{}, placeHolder Placeholder) (string, []interface{}, bool) {
	word := regexp.MustCompile(`\&\{\w+[\.]?\w+\}`)
	hasAnd := false
	vals := []interface{}{}
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		hasAnd = true
		fullKey := s[2 : len(s)-1]
		val, party := andSymbols(input, fullKey, placeHolder)
		if party != "" {
			vals = append(vals, val)
		}
		return party
	})

	return sql, vals, hasAnd
}

func andSymbols(input map[string]interface{}, fullKey string, placeHolder Placeholder) (value interface{}, party string) {
	propName := GetPropName(fullKey)
	value = input[propName]
	if !IsNil(value) {
		return value, fmt.Sprintf(" and %s=%s", fullKey, placeHolder())
	}
	return value, ""
}
