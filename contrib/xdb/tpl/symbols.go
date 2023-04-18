package tpl

import (
	"database/sql"
	"fmt"
	"reflect"
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
			item.Values = append(item.Values, nil)
		}
		return argName
	}

	defaultSymbols["&"] = func(input DBParam, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		argName, value := input.Get(propName, item.Placeholder)
		item.HasAndOper = true
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" and %s=%s", fullKey, argName)
		}
		return ""
	}
	defaultSymbols["|"] = func(input DBParam, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		argName, value := input.Get(propName, item.Placeholder)
		item.HasOrOper = true
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" or %s=%s", fullKey, argName)
		}
		return ""
	}

}

func IsNil(input interface{}) bool {
	if input == nil {
		return true
	}

	if arg, ok := input.(sql.NamedArg); ok {
		input = arg.Value
	}
	if arg, ok := input.(*sql.NamedArg); ok {
		input = arg.Value
	}

	if fmt.Sprintf("%v", input) == "" {
		return true
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Ptr {
		return rv.IsNil()
	}
	return false
}
