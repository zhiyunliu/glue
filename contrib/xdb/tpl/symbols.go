package tpl

import (
	"database/sql"
	"fmt"
)

var defaultSymbols Symbols

func init() {
	defaultSymbols = make(Symbols)
	defaultSymbols["@"] = func(input DBParam, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		argName, value := input.Get(propName, item.Placeholder)
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
		} else {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, sql.Named(argName, nil))
		}
		return argName
	}

	defaultSymbols["&"] = func(input DBParam, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		argName, value := input.Get(propName, item.Placeholder)
		if !IsNil(value.Value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" and %s=%s", fullKey, argName)
		}
		return ""
	}
	defaultSymbols["|"] = func(input DBParam, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		argName, value := input.Get(propName, item.Placeholder)
		if !IsNil(value.Value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" or %s=%s", fullKey, argName)
		}
		return ""
	}

}
