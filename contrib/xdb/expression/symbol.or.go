package expression

import "github.com/zhiyunliu/glue/xdb"

type orSymbols struct{}

func (s *orSymbols) Name() string {
	return xdb.SymbolOr
}

func (s *orSymbols) Concat() string {
	return "or"
}
func (s *orSymbols) Callback(item *xdb.SqlScene, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.HasDynamicOr = true

	propName := valuer.GetPropName()

	argName, value, _ := input.Get(propName, item.Placeholder)

	if !xdb.IsNil(value) {
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, value)
		return valuer.Build(input, argName)
	}
	return "", nil
}
