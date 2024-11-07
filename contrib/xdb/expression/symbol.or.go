package expression

import "github.com/zhiyunliu/glue/xdb"

type orSymbols struct{}

func (s *orSymbols) Name() string {
	return xdb.SymbolOr
}

func (s *orSymbols) Concat() string {
	return "or"
}
func (s *orSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.SetDynamic(xdb.DynamicOr)
	propName := valuer.GetPropName()
	argName, value, _ := input.Get(propName, item.GetPlaceholder())

	if !xdb.IsNil(value) {
		item.AppendExpr(propName, value)
		return valuer.Build(input, argName)
	}
	return "", nil
}
