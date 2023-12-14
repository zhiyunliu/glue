package sqlserver

import (
	"fmt"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

func newMssqlSymbols() tpl.Symbols {

	symbols := make(tpl.Symbols)
	symbols["@"] = func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissParamError) {
		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[propName]; ok {
			return ph, nil
		}
		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return argName, err
		}
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, value)

		item.NameCache[propName] = argName
		return argName, nil
	}

	symbols["&"] = func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissParamError) {
		item.HasAndOper = true

		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[propName]; ok {
			return fmt.Sprintf("and %s=%s ", fullKey, ph), nil
		}
		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return "", err
		}
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return fmt.Sprintf("and %s=%s ", fullKey, argName), nil
		}
		return "", nil
	}
	symbols["|"] = func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissParamError) {
		item.HasOrOper = true

		propName := tpl.GetPropName(fullKey)
		if ph, ok := item.NameCache[propName]; ok {
			return fmt.Sprintf("or %s=%s ", fullKey, ph), nil
		}
		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return argName, err
		}
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return fmt.Sprintf("or %s=%s ", fullKey, argName), nil
		}
		return "", nil
	}
	return symbols
}
