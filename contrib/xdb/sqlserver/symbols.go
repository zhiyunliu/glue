package sqlserver

import (
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

func newMssqlSymbols(operMap tpl.OperatorMap) tpl.SymbolMap {

	symbols := tpl.NewSymbolMap(operMap)
	symbols.StoreSymbol(tpl.SymbolAt, func(symbolMap tpl.SymbolMap, input xdb.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissError) {

		matcher := tpl.GetPropMatchValuer(fullKey, item.PropOpts)
		if matcher == nil {
			return "", xdb.NewMissPropError(fullKey)
		}
		propName := matcher.GetPropName()

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
	})

	symbols.StoreSymbol(tpl.SymbolAnd, func(symbolMap tpl.SymbolMap, input xdb.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissError) {
		item.HasAndOper = true

		matcher := tpl.GetPropMatchValuer(fullKey, item.PropOpts)
		if matcher == nil {
			return "", xdb.NewMissPropError(fullKey)
		}
		propName := matcher.GetPropName()

		if ph, ok := item.NameCache[propName]; ok {
			return matcher.Build(tpl.SymbolAnd, input, ph)
		}
		argName, value, _ := input.Get(propName, item.Placeholder)
		if !xdb.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return matcher.Build(tpl.SymbolAnd, input, argName)

		}
		return "", nil
	})

	symbols.StoreSymbol(tpl.SymbolOr, func(symbolMap tpl.SymbolMap, input xdb.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissError) {
		item.HasOrOper = true

		matcher := tpl.GetPropMatchValuer(fullKey, item.PropOpts)
		if matcher == nil {
			return "", xdb.NewMissPropError(fullKey)
		}
		propName := matcher.GetPropName()

		if ph, ok := item.NameCache[propName]; ok {
			return matcher.Build(tpl.SymbolOr, input, ph)
		}

		argName, value, _ := input.Get(propName, item.Placeholder)
		if !xdb.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return matcher.Build(tpl.SymbolOr, input, argName)
		}
		return "", nil
	})

	return symbols
}
