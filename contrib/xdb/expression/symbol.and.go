package expression

import "github.com/zhiyunliu/glue/xdb"

type andSymbols struct{}

func (s *andSymbols) Name() string {
	return xdb.SymbolAnd
}

func (s *andSymbols) Concat() string {
	return "and"
}

func (s *andSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.SetDynamic(xdb.DynamicAnd)

	propName := valuer.GetPropName()
	argName, value, _ := input.Get(propName, item.GetPlaceholder())
	if !xdb.IsNil(value) {
		item.AppendExpr(propName, value)
		return valuer.Build(input, argName)
	}
	return "", nil
}
