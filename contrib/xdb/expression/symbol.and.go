package expression

import "github.com/zhiyunliu/glue/xdb"

type andSymbols struct{}

func (s *andSymbols) Name() string {
	return xdb.SymbolAnd
}

func (s *andSymbols) Concat() string {
	return "and"
}

func (s *andSymbols) Callback(item *xdb.SqlScene, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.HasDynamicAnd = true

	propName := valuer.GetPropName()

	argName, value, _ := input.Get(propName, item.Placeholder)
	if !xdb.IsNil(value) {
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, value)
		return valuer.Build(input, argName)
	}
	return "", nil
}
