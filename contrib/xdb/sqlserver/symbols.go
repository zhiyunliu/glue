package sqlserver

import (
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

func newMssqlSymbols(operMap tpl.OperatorMap) tpl.SymbolMap {

	symbols := tpl.NewSymbolMap(operMap)
	symbols.LoadOrStore(tpl.SymbolAt, func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissError) {
		_, propName, _ := tpl.GetPropName(fullKey)
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

	symbols.LoadOrStore(tpl.SymbolAnd, func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissError) {
		item.HasAndOper = true

		fullField, propName, oper := tpl.GetPropName(fullKey)
		opercall, ok := operMap.Load(oper)
		if !ok {
			return "", xdb.NewMissOperError(oper)
		}

		if ph, ok := item.NameCache[propName]; ok {
			return opercall(tpl.SymbolAnd, fullField, ph), nil
			//return fmt.Sprintf("and %s=%s ", fullKey, ph), nil
		}
		argName, value, _ := input.Get(propName, item.Placeholder)
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName
			return opercall(tpl.SymbolAnd, fullField, argName), nil
			//return fmt.Sprintf("and %s=%s ", fullKey, argName), nil
		}
		return "", nil
	})

	symbols.LoadOrStore(tpl.SymbolOr, func(input tpl.DBParam, fullKey string, item *tpl.ReplaceItem) (string, xdb.MissError) {
		item.HasOrOper = true

		fullField, propName, oper := tpl.GetPropName(fullKey)
		opercall, ok := operMap.Load(oper)
		if !ok {
			return "", xdb.NewMissOperError(oper)
		}
		if ph, ok := item.NameCache[propName]; ok {
			return opercall(tpl.SymbolOr, fullField, ph), nil
			//return fmt.Sprintf("or %s=%s ", fullKey, ph), nil
		}
		argName, value, _ := input.Get(propName, item.Placeholder)
		if !tpl.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			item.NameCache[propName] = argName

			return opercall(tpl.SymbolOr, fullField, argName), nil
			//return fmt.Sprintf("or %s=%s ", fullKey, argName), nil
		}
		return "", nil
	})

	return symbols
}
