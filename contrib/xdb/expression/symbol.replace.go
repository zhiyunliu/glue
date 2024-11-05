package expression

import "github.com/zhiyunliu/glue/xdb"

type replaceSymbols struct{}

func (s *replaceSymbols) Name() string {
	return xdb.SymbolAnd
}

func (s *replaceSymbols) Concat() string {
	return ""
}
func (s *replaceSymbols) Callback(item *xdb.SqlScene, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.HasDynamicReplace = true

	propName := valuer.GetPropName()

	argName, value, _ := input.Get(propName, item.Placeholder)

	if !xdb.IsNil(value) {
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, value)
		return valuer.Build(input, argName)
	}
	return "", nil
}
