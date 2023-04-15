package sqlserver

import (
	"fmt"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
)

func newMssqlSymbols() tpl.Symbols {

	symbols := make(tpl.Symbols)
	symbols["@"] = func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[propName]; ok {
			return ph
		}
		argName, value := input.Get(propName, item.Placeholder)
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, value)

		item.NameCache[propName] = argName
		return argName
	}

	symbols["&"] = func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[propName]; ok {
			return fmt.Sprintf("and %s=%s ", fullKey, ph)
		}
		argName, value := input.Get(propName, item.Placeholder)
		if !tpl.IsNil(value.Value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return fmt.Sprintf("and %s=%s ", fullKey, argName)
		}
		return ""
	}
	symbols["|"] = func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) string {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[propName]; ok {
			return fmt.Sprintf("or %s=%s ", fullKey, ph)
		}
		argName, value := input.Get(propName, item.Placeholder)
		if !tpl.IsNil(value.Value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return fmt.Sprintf("or %s=%s ", fullKey, argName)
		}
		return ""
	}
	return symbols
}
