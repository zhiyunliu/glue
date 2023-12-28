package tpl

import (
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
)

// 处理替换符合
func handleRelaceSymbols(tpl string, input map[string]interface{}, ph Placeholder) (string, bool, error) {
	word := GetPatternRegexp(ReplacePattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ph,
	}
	hasReplace := false
	var outerrs []xdb.MissParamError
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		hasReplace = true
		fullKey := s[2 : len(s)-1]
		tmpv, err := replaceSymbols(input, fullKey, item)
		if err != nil {
			outerrs = append(outerrs, err)
		}
		return tmpv
	})
	if len(outerrs) > 0 {
		return sql, hasReplace, xdb.NewMissParamsError(outerrs...)
	}
	return sql, hasReplace, nil
}

func replaceSymbols(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissParamError) {
	propName := GetPropName(fullKey)
	value, err := input.GetVal(propName)
	if err != nil {
		return "", err
	}
	if !IsNil(value) {
		return fmt.Sprintf("%v", value), nil
	}
	return "", nil
}
