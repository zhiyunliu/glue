package tpl

import (
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
)

// 处理替换符合
func handleRelaceSymbols(tpl string, input map[string]interface{}, ph xdb.Placeholder) (string, bool, error) {
	word := GetPatternRegexp(ReplacePattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: ph,
		PropOpts: &xdb.PropOptions{
			UseCache: true,
		},
	}
	hasReplace := false
	var outerrs []xdb.MissError
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
		return sql, hasReplace, xdb.NewMissListError(outerrs...)
	}
	return sql, hasReplace, nil
}

func replaceSymbols(input xdb.DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
	//	_, propName, _ := GetPropName(fullKey, item.PropOpts)

	matcher := GetPropMatchValuer(fullKey, item.PropOpts)
	if matcher == nil {
		return "", xdb.NewMissPropError(fullKey)
	}
	propName := matcher.GetPropName()

	value, err := input.GetVal(propName)
	if err != nil {
		return "", err
	}
	if !xdb.IsNil(value) {
		return fmt.Sprintf("%v", value), nil
	}
	return "", nil
}
